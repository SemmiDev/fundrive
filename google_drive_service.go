package fundrive

import (
	"context"
	"fmt"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
	"io"
	"math"
	"net/http"
	"strings"
	"time"
)

type CreateFolderRequest struct {
	UserID      string   `json:"user_id"`
	Email       string   `json:"email"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Parents     []string `json:"parents"`
}

func (service *GoogleDriveService) CreateFolder(ctx context.Context, req *CreateFolderRequest) (*drive.File, error) {
	newDriveServiceReq := newDriveServiceRequest{
		UserID: req.UserID,
		Email:  req.Email,
	}

	srv, err := service.newDriveService(ctx, &newDriveServiceReq)

	if err != nil {
		return nil, fmt.Errorf("error creating google drive service: %w", err)
	}

	folder := &drive.File{
		MimeType:    MimeTypeFolder,
		Name:        req.Name,
		Description: req.Description,
		Parents:     req.Parents,
	}

	file, err := srv.Files.Create(folder).Do()
	if err != nil {
		return nil, fmt.Errorf("error creating folder: %w", err)
	}

	return file, nil
}

type ListFoldersRequest struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	PageSize  int64  `json:"page_size"`
	PageToken string `json:"page_token"`
}

func (l *ListFoldersRequest) HasPageToken() bool {
	return l.PageToken != ""
}

func (l *ListFoldersRequest) HasPageSize() bool {
	return l.PageSize != 0
}

func (service *GoogleDriveService) ListFolders(ctx context.Context, req *ListFoldersRequest) ([]*drive.File, string, error) {
	newDriveServiceReq := newDriveServiceRequest{
		UserID: req.UserID,
		Email:  req.Email,
	}

	srv, err := service.newDriveService(ctx, &newDriveServiceReq)
	if err != nil {
		return nil, "", fmt.Errorf("error creating google drive service: %w", err)
	}

	q := fmt.Sprintf("mimeType = '%s'", MimeTypeFolder)

	request := srv.Files.List().Q(q).
		Spaces("drive").
		Fields("nextPageToken, files(id, name, mimeType, parents)")

	if req.HasPageToken() {
		request = request.PageToken(req.PageToken)
	}

	if req.HasPageSize() {
		request = request.PageSize(req.PageSize)
	}

	response, err := request.Do()
	if err != nil {
		return nil, "", fmt.Errorf("error listing folders: %w", err)
	}

	return response.Files, response.NextPageToken, nil
}

type UploadFileRequest struct {
	UserID   string    `json:"user_id"`
	Email    string    `json:"email"`
	FileName string    `json:"file_name"`
	MimeType string    `json:"mime_type"`
	FileData io.Reader `json:"file_data"`
	Parents  []string  `json:"parents"`
}

func (u *UploadFileRequest) Sanitize() {
	u.FileName = strings.ReplaceAll(u.FileName, "/", "_")
}

func (service *GoogleDriveService) UploadFile(ctx context.Context, req *UploadFileRequest) (*drive.File, error) {
	newDriveServiceReq := newDriveServiceRequest{
		UserID: req.UserID,
		Email:  req.Email,
	}

	srv, err := service.newDriveService(ctx, &newDriveServiceReq)

	if err != nil {
		return nil, fmt.Errorf("error creating google drive service: %w", err)
	}

	// sanitize the filename
	req.Sanitize()

	file := &drive.File{
		Name:    req.FileName,
		Parents: req.Parents,
		//MimeType: req.MimeType, // auto detect mime type
	}

	// create permission
	permission := &drive.Permission{
		Type: "anyone",
		Role: "reader",
	}

	uploadedFile, err := srv.Files.
		Create(file).
		Media(req.FileData).
		Do()
	if err != nil {
		return nil, err
	}

	// save permission
	_, err = srv.Permissions.Create(uploadedFile.Id, permission).Do()
	if err != nil {
		return nil, err
	}

	return uploadedFile, nil
}

type ListFilesInFolderRequest struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	FolderID string `json:"folder_id"`
}

func (service *GoogleDriveService) ListFilesInFolder(ctx context.Context, req *ListFilesInFolderRequest) ([]*drive.File, error) {
	newDriveServiceReq := newDriveServiceRequest{
		UserID: req.UserID,
		Email:  req.Email,
	}

	srv, err := service.newDriveService(ctx, &newDriveServiceReq)

	if err != nil {
		return nil, fmt.Errorf("error creating google drive service: %w", err)
	}

	q := fmt.Sprintf("mimeType != '%s' and trashed = false and '%s' in parents", MimeTypeFolder, req.FolderID)

	request := srv.Files.List().Q(q).Spaces("drive").
		Fields("nextPageToken, files(id, name, mimeType)").
		Corpora("user") // owned by the user.

	response, err := request.Do()
	if err != nil {
		return nil, fmt.Errorf("error listing files: %w", err)
	}

	return response.Files, nil
}

type DeleteResourceRequest struct {
	UserID     string `json:"user_id"`
	Email      string `json:"email"`
	ResourceID string `json:"resource_id"`
}

func (service *GoogleDriveService) Delete(ctx context.Context, req *DeleteResourceRequest) error {
	newDriveServiceReq := newDriveServiceRequest{
		UserID: req.UserID,
		Email:  req.Email,
	}

	srv, err := service.newDriveService(ctx, &newDriveServiceReq)

	if err != nil {
		return fmt.Errorf("error creating google drive service: %w", err)
	}

	return srv.Files.Delete(req.ResourceID).Do()
}

type GetFileRequest struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	FileID string `json:"file_id"`
}

func (service *GoogleDriveService) GetFile(ctx context.Context, req *GetFileRequest) (*drive.File, error) {
	newDriveServiceReq := newDriveServiceRequest{
		UserID: req.UserID,
		Email:  req.Email,
	}

	srv, err := service.newDriveService(ctx, &newDriveServiceReq)

	if err != nil {
		return nil, fmt.Errorf("error creating google drive service: %w", err)
	}

	return srv.Files.Get(req.FileID).Do()
}

func (service *GoogleDriveService) GetFileWithURL(ctx context.Context, req *GetFileRequest) (*drive.File, error) {
	newDriveServiceReq := newDriveServiceRequest{
		UserID: req.UserID,
		Email:  req.Email,
	}

	srv, err := service.newDriveService(ctx, &newDriveServiceReq)

	if err != nil {
		return nil, fmt.Errorf("error creating google drive service: %w", err)
	}

	return srv.Files.Get(req.FileID).Fields("webViewLink").Do()
}

type DownloadFileRequest struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	FileID string `json:"file_id"`
}

type DownloadFileResponse struct {
	FileName         string `json:"file_name"`
	FileExt          string `json:"file_ext"`
	OriginalFileName string `json:"original_file_name"`
	MimeType         string `json:"mime_type"`
	Response         *http.Response
}

func (service *GoogleDriveService) DownloadFile(ctx context.Context, req *DownloadFileRequest) (*DownloadFileResponse, error) {
	newDriveServiceReq := newDriveServiceRequest{
		UserID: req.UserID,
		Email:  req.Email,
	}

	srv, err := service.newDriveService(ctx, &newDriveServiceReq)

	if err != nil {
		return nil, fmt.Errorf("error creating google drive service: %w", err)
	}

	fileInfo, err := srv.Files.Get(req.FileID).Do(googleapi.QueryParameter("alt", "media"))

	file, err := srv.Files.Get(req.FileID).Download(
		googleapi.QueryParameter("alt", "media"),
	)

	if err != nil {
		return nil, err
	}

	return &DownloadFileResponse{
		FileName:         fileInfo.Name,
		FileExt:          fileInfo.FileExtension,
		OriginalFileName: fileInfo.OriginalFilename,
		MimeType:         fileInfo.MimeType,
		Response:         file,
	}, nil
}

type StorageInfo struct {
	KuotaTerpakaiInGB float64 `json:"kuota_terpakai_in_gb"`
	KuotaTotalInGB    float64 `json:"kuota_total_in_gb"`
	KuotaSisaInGB     float64 `json:"kuota_sisa_in_gb"`

	KuotaTerpakaiInMB float64 `json:"kuota_terpakai_in_mb"`
	KuotaTotalInMB    float64 `json:"kuota_total_in_mb"`
	KuotaSisaInMB     float64 `json:"kuota_sisa_in_mb"`

	KuotaTerpakaiInKB float64 `json:"kuota_terpakai_in_kb"`
	KuotaTotalInKB    float64 `json:"kuota_total_in_kb"`
	KuotaSisaInKB     float64 `json:"kuota_sisa_in_kb"`

	KuotaTerpakaiInBytes float64 `json:"kuota_terpakai_in_bytes"`
	KuotaTotalInBytes    float64 `json:"kuota_total_in_bytes"`
	KuotaSisaInBytes     float64 `json:"kuota_sisa_in_bytes"`
}

func (s *StorageInfo) FormatTwoDigits() {
	s.KuotaTerpakaiInGB = math.Round(s.KuotaTerpakaiInGB*100) / 100
	s.KuotaTotalInGB = math.Round(s.KuotaTotalInGB*100) / 100
	s.KuotaSisaInGB = math.Round(s.KuotaSisaInKB*100) / 100

	s.KuotaTerpakaiInMB = math.Round(s.KuotaTerpakaiInMB*100) / 100
	s.KuotaTotalInMB = math.Round(s.KuotaTotalInMB*100) / 100
	s.KuotaSisaInMB = math.Round(s.KuotaSisaInMB*100) / 100

	s.KuotaTerpakaiInKB = math.Round(s.KuotaTerpakaiInKB*100) / 100
	s.KuotaTotalInKB = math.Round(s.KuotaTotalInKB*100) / 100
	s.KuotaSisaInKB = math.Round(s.KuotaSisaInKB*100) / 100

	s.KuotaTerpakaiInBytes = math.Round(s.KuotaTerpakaiInBytes*100) / 100
	s.KuotaTotalInBytes = math.Round(s.KuotaTotalInBytes*100) / 100
	s.KuotaSisaInBytes = math.Round(s.KuotaSisaInBytes*100) / 100
}

type GetStorageInfoRequest struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

func (service *GoogleDriveService) GetStorageInfo(ctx context.Context, req *GetStorageInfoRequest) (*StorageInfo, error) {
	newDriveServiceReq := newDriveServiceRequest{
		UserID: req.UserID,
		Email:  req.Email,
	}

	srv, err := service.newDriveService(ctx, &newDriveServiceReq)

	if err != nil {
		return nil, fmt.Errorf("error creating google drive service: %w", err)
	}

	about := srv.About.Get()

	aboutResult, err := about.Fields("storageQuota").Do()
	if err != nil {
		return nil, fmt.Errorf("unable to get About info: %v", err)
	}

	quota := aboutResult.StorageQuota
	usageInBytes := quota.Usage
	limitInBytes := quota.Limit
	remainingInBytes := limitInBytes - usageInBytes

	usageInGB := float64(usageInBytes) / 1024 / 1024 / 1024
	limitInGB := float64(limitInBytes) / 1024 / 1024 / 1024
	remainingInGB := float64(remainingInBytes) / 1024 / 1024 / 1024

	usageInMB := float64(usageInBytes) / 1024 / 1024
	limitInMB := float64(limitInBytes) / 1024 / 1024
	remainingInMB := float64(remainingInBytes) / 1024 / 1024

	usageInKB := float64(usageInBytes) / 1024
	limitInKB := float64(limitInBytes) / 1024
	remainingInKB := float64(remainingInBytes) / 1024

	return &StorageInfo{
		KuotaTerpakaiInGB: usageInGB,
		KuotaTotalInGB:    limitInGB,
		KuotaSisaInGB:     remainingInGB,

		KuotaTerpakaiInMB: usageInMB,
		KuotaTotalInMB:    limitInMB,
		KuotaSisaInMB:     remainingInMB,

		KuotaTerpakaiInKB: usageInKB,
		KuotaTotalInKB:    limitInKB,
		KuotaSisaInKB:     remainingInKB,
	}, nil
}

// RenameResourceRequest represents the request to rename a file or folder
type RenameResourceRequest struct {
	UserID     string `json:"user_id" validate:"required"`
	Email      string `json:"email" validate:"required"`
	ResourceID string `json:"resource_id" validate:"required"`
	NewName    string `json:"new_name" validate:"required"`
}

func (r *RenameResourceRequest) Validate() error {
	if r.UserID == "" {
		return fmt.Errorf("user ID is required")
	}
	if r.ResourceID == "" {
		return fmt.Errorf("resource ID is required")
	}
	if r.NewName == "" {
		return fmt.Errorf("new name is required")
	}
	return nil
}

func (service *GoogleDriveService) RenameResource(ctx context.Context, req *RenameResourceRequest) (*drive.File, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid rename request: %w", err)
	}

	// Get Drive service
	newDriveServiceReq := newDriveServiceRequest{
		UserID: req.UserID,
		Email:  req.Email,
	}

	srv, err := service.newDriveService(ctx, &newDriveServiceReq)

	if err != nil {
		return nil, fmt.Errorf("error creating google drive service: %w", err)
	}

	// Check if resource exists
	_, err = srv.Files.Get(req.ResourceID).Fields("id, name, mimeType").Do()
	if err != nil {
		return nil, fmt.Errorf("error getting resource: %w", err)
	}

	// Create update request with new name
	updateFile := &drive.File{
		Name: req.NewName,
	}

	// Perform update
	updatedFile, err := srv.Files.Update(req.ResourceID, updateFile).
		Fields("id, name, mimeType, modifiedTime").
		Do()
	if err != nil {
		return nil, fmt.Errorf("error renaming resource: %w", err)
	}

	return updatedFile, nil
}

type MoveResourceRequest struct {
	UserID       string   `json:"user_id" validate:"required"`
	Email        string   `json:"email" validate:"required"`
	ResourceID   string   `json:"resource_id" validate:"required"`
	NewParentID  string   `json:"new_parent_id" validate:"required"`
	OldParentIDs []string `json:"old_parent_ids,omitempty"`
}

func (service *GoogleDriveService) MoveResource(ctx context.Context, req *MoveResourceRequest) (*drive.File, error) {
	newDriveServiceReq := newDriveServiceRequest{
		UserID: req.UserID,
		Email:  req.Email,
	}

	srv, err := service.newDriveService(ctx, &newDriveServiceReq)

	if err != nil {
		return nil, fmt.Errorf("error creating google drive service: %w", err)
	}

	// Prepare the file with new parent
	file := &drive.File{}
	if len(req.OldParentIDs) > 0 {
		// Remove old parents if specified
		for _, oldParentID := range req.OldParentIDs {
			file.Parents = append(file.Parents, oldParentID)
		}
	}

	// Move the file to new parent
	updatedFile, err := srv.Files.Update(req.ResourceID, file).
		AddParents(req.NewParentID).
		RemoveParents(strings.Join(req.OldParentIDs, ",")).
		Fields("id, name, parents, mimeType").
		Do()

	if err != nil {
		return nil, fmt.Errorf("error moving resource: %w", err)
	}

	return updatedFile, nil
}

type CopyResourceRequest struct {
	UserID              string `json:"user_id" validate:"required"`
	Email               string `json:"email" validate:"required"`
	ResourceID          string `json:"resource_id" validate:"required"`
	DestinationParentID string `json:"destination_parent_id" validate:"required"`
	NewName             string `json:"new_name,omitempty"`
}

func (service *GoogleDriveService) CopyResource(ctx context.Context, req *CopyResourceRequest) (*drive.File, error) {
	newDriveServiceReq := newDriveServiceRequest{
		UserID: req.UserID,
		Email:  req.Email,
	}

	srv, err := service.newDriveService(ctx, &newDriveServiceReq)

	if err != nil {
		return nil, fmt.Errorf("error creating google drive service: %w", err)
	}

	// Prepare copy metadata
	copyFile := &drive.File{
		Name:    req.NewName,
		Parents: []string{req.DestinationParentID},
	}

	// Perform copy operation
	copiedFile, err := srv.Files.Copy(req.ResourceID, copyFile).
		Fields("id, name, mimeType, parents").
		Do()

	if err != nil {
		return nil, fmt.Errorf("error copying resource: %w", err)
	}

	return copiedFile, nil
}

type SearchResourcesRequest struct {
	UserID    string `json:"user_id" validate:"required"`
	Email     string `json:"email" validate:"required"`
	Query     string `json:"query" validate:"required"`
	PageToken string `json:"page_token,omitempty"`
	PageSize  int64  `json:"page_size,omitempty"`
	MimeType  string `json:"mime_type,omitempty"`
	Trashed   bool   `json:"trashed,omitempty"`
}

func (service *GoogleDriveService) SearchResources(ctx context.Context, req *SearchResourcesRequest) ([]*drive.File, string, error) {
	newDriveServiceReq := newDriveServiceRequest{
		UserID: req.UserID,
		Email:  req.Email,
	}

	srv, err := service.newDriveService(ctx, &newDriveServiceReq)

	if err != nil {
		return nil, "", fmt.Errorf("error creating google drive service: %w", err)
	}

	// Build search query
	queryParts := []string{fmt.Sprintf("fullText contains '%s'", req.Query)}
	if req.MimeType != "" {
		queryParts = append(queryParts, fmt.Sprintf("mimeType = '%s'", req.MimeType))
	}
	queryParts = append(queryParts, fmt.Sprintf("trashed = %t", req.Trashed))

	// Create list request
	listReq := srv.Files.List().
		Q(strings.Join(queryParts, " and ")).
		Fields("nextPageToken, files(id, name, mimeType, parents, size, createdTime, modifiedTime)")

	if req.PageToken != "" {
		listReq = listReq.PageToken(req.PageToken)
	}
	if req.PageSize > 0 {
		listReq = listReq.PageSize(req.PageSize)
	}

	// Execute search
	result, err := listReq.Do()
	if err != nil {
		return nil, "", fmt.Errorf("error searching resources: %w", err)
	}

	return result.Files, result.NextPageToken, nil
}

type UpdatePermissionRequest struct {
	UserID       string `json:"user_id" validate:"required"`
	Email        string `json:"email" validate:"required"`
	ResourceID   string `json:"resource_id" validate:"required"`
	EmailAddress string `json:"email_address" validate:"required,email"`
	Role         string `json:"role" validate:"required,oneof=reader writer owner"`
	Type         string `json:"type" validate:"required,oneof=user group domain anyone"`
	NotifyEmail  bool   `json:"notify_email"`
}

func (service *GoogleDriveService) UpdatePermissions(ctx context.Context, req *UpdatePermissionRequest) error {
	newDriveServiceReq := newDriveServiceRequest{
		UserID: req.UserID,
		Email:  req.Email,
	}

	srv, err := service.newDriveService(ctx, &newDriveServiceReq)

	if err != nil {
		return fmt.Errorf("error creating google drive service: %w", err)
	}

	permission := &drive.Permission{
		EmailAddress: req.EmailAddress,
		Role:         req.Role,
		Type:         req.Type,
	}

	// Create permission
	_, err = srv.Permissions.Create(req.ResourceID, permission).
		SendNotificationEmail(req.NotifyEmail).
		Do()

	if err != nil {
		return fmt.Errorf("error updating permissions: %w", err)
	}

	return nil
}

type ResourceMetadata struct {
	ID           string              `json:"id"`
	Name         string              `json:"name"`
	MimeType     string              `json:"mime_type"`
	Size         int64               `json:"size"`
	CreatedTime  time.Time           `json:"created_time"`
	ModifiedTime time.Time           `json:"modified_time"`
	ViewedTime   time.Time           `json:"viewed_time"`
	Owners       []string            `json:"owners"`
	SharedWithMe bool                `json:"shared_with_me"`
	Starred      bool                `json:"starred"`
	Trashed      bool                `json:"trashed"`
	WebViewLink  string              `json:"web_view_link"`
	IconLink     string              `json:"icon_link"`
	Permissions  []*drive.Permission `json:"permissions"`
}

type GetMetadataRequest struct {
	UserID     string `json:"user_id" validate:"required"`
	Email      string `json:"email" validate:"required"`
	ResourceID string `json:"resource_id" validate:"required"`
}

func (service *GoogleDriveService) GetResourceMetadata(ctx context.Context, req *GetMetadataRequest) (*ResourceMetadata, error) {
	newDriveServiceReq := newDriveServiceRequest{
		UserID: req.UserID,
		Email:  req.Email,
	}

	srv, err := service.newDriveService(ctx, &newDriveServiceReq)

	if err != nil {
		return nil, fmt.Errorf("error creating google drive service: %w", err)
	}
	file, err := srv.Files.Get(req.ResourceID).
		Fields("id, name, mimeType, size, createdTime, modifiedTime, viewedByMeTime, owners, sharedWithMeTime, starred, trashed, webViewLink, iconLink, permissions").
		Do()

	if err != nil {
		return nil, fmt.Errorf("error getting resource metadata: %w", err)
	}

	owners := make([]string, 0)
	for _, owner := range file.Owners {
		owners = append(owners, owner.EmailAddress)
	}

	createdTime, err := time.Parse(time.RFC3339, file.CreatedTime)
	if err != nil {
		return nil, fmt.Errorf("error parsing created time: %w", err)
	}

	modifiedTime, err := time.Parse(time.RFC3339, file.ModifiedTime)
	if err != nil {
		return nil, fmt.Errorf("error parsing modified time: %w", err)
	}

	viewedTime, err := time.Parse(time.RFC3339, file.ViewedByMeTime)
	if err != nil {
		return nil, fmt.Errorf("error parsing viewed time: %w", err)
	}

	return &ResourceMetadata{
		ID:           file.Id,
		Name:         file.Name,
		MimeType:     file.MimeType,
		Size:         file.Size,
		CreatedTime:  createdTime,
		ModifiedTime: modifiedTime,
		ViewedTime:   viewedTime,
		Owners:       owners,
		SharedWithMe: file.Shared,
		Starred:      file.Starred,
		Trashed:      file.Trashed,
		WebViewLink:  file.WebViewLink,
		IconLink:     file.IconLink,
		Permissions:  file.Permissions,
	}, nil
}

type RestoreRequest struct {
	UserID     string `json:"user_id" validate:"required"`
	Email      string `json:"email" validate:"required"`
	ResourceID string `json:"resource_id" validate:"required"`
}

func (service *GoogleDriveService) RestoreFromTrash(ctx context.Context, req *RestoreRequest) error {
	newDriveServiceReq := newDriveServiceRequest{
		UserID: req.UserID,
		Email:  req.Email,
	}

	srv, err := service.newDriveService(ctx, &newDriveServiceReq)

	if err != nil {
		return fmt.Errorf("error creating google drive service: %w", err)
	}

	// Untrash the file
	_, err = srv.Files.Update(req.ResourceID, &drive.File{
		Trashed: false,
	}).Do()

	if err != nil {
		return fmt.Errorf("error restoring resource from trash: %w", err)
	}

	return nil
}

type EmptyTrashRequest struct {
	UserID string `json:"user_id" validate:"required"`
	Email  string `json:"email" validate:"required"`
}

func (service *GoogleDriveService) EmptyTrash(ctx context.Context, req *EmptyTrashRequest) error {
	newDriveServiceReq := newDriveServiceRequest{
		UserID: req.UserID,
		Email:  req.Email,
	}

	srv, err := service.newDriveService(ctx, &newDriveServiceReq)

	if err != nil {
		return fmt.Errorf("error creating google drive service: %w", err)
	}

	err = srv.Files.EmptyTrash().Do()
	if err != nil {
		return fmt.Errorf("error emptying trash: %w", err)
	}

	return nil
}

type ExportFileRequest struct {
	UserID   string `json:"user_id" validate:"required"`
	Email    string `json:"email" validate:"required"`
	FileID   string `json:"file_id" validate:"required"`
	MimeType string `json:"mime_type" validate:"required"`
}

type ExportFileResponse struct {
	Content    []byte    `json:"content"`
	MimeType   string    `json:"mime_type"`
	ExportedAt time.Time `json:"exported_at"`
}

func (service *GoogleDriveService) ExportFile(ctx context.Context, req *ExportFileRequest) (*ExportFileResponse, error) {
	newDriveServiceReq := newDriveServiceRequest{
		UserID: req.UserID,
		Email:  req.Email,
	}

	srv, err := service.newDriveService(ctx, &newDriveServiceReq)

	if err != nil {
		return nil, fmt.Errorf("error creating google drive service: %w", err)
	}

	response, err := srv.Files.Export(req.FileID, req.MimeType).Download()
	if err != nil {
		return nil, fmt.Errorf("error exporting file: %w", err)
	}
	defer response.Body.Close()

	content, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading exported content: %w", err)
	}

	return &ExportFileResponse{
		Content:    content,
		MimeType:   req.MimeType,
		ExportedAt: time.Now(),
	}, nil
}
