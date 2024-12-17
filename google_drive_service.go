package fundrive

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

type CreateFolderRequest struct {
	UserID      string     `json:"user_id" validate:"required"`
	Email       string     `json:"email" validate:"required,email"`
	Name        string     `json:"name" validate:"required"`
	Description string     `json:"description"`
	Permission  Permission `json:"permission"`
	Parents     []string   `json:"parents"`
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

	request := &drive.File{
		MimeType:    MimeTypeFolder,
		Name:        req.Name,
		Description: req.Description,
		Parents:     req.Parents,
	}

	response, err := srv.Files.Create(request).Do()
	if err != nil {
		return nil, fmt.Errorf("error creating folder: %w", err)
	}

	permission := getPermission(req.Permission)
	_, err = srv.Permissions.Create(response.Id, permission).Do()
	if err != nil {
		return nil, err
	}

	return response, nil
}

type ListFoldersRequest struct {
	UserID    string `json:"user_id" validate:"required"`
	Email     string `json:"email" validate:"required"`
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
	UserID     string     `json:"user_id" validate:"required"`
	Email      string     `json:"email" validate:"required,email"`
	FileName   string     `json:"file_name" validate:"required"`
	MimeType   string     `json:"mime_type"`
	FileData   io.Reader  `json:"file_data"`
	Permission Permission `json:"permission"`
	Parents    []string   `json:"parents"`
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
	}

	response, err := srv.Files.
		Create(file).
		Media(req.FileData).
		Do()
	if err != nil {
		return nil, err
	}

	permission := getPermission(req.Permission)
	_, err = srv.Permissions.Create(response.Id, permission).Do()
	if err != nil {
		return nil, err
	}

	return response, nil
}

type ListFilesInFolderRequest struct {
	UserID   string `json:"user_id" validate:"required"`
	Email    string `json:"email" validate:"required"`
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
	UserID string `json:"user_id" validate:"required"`
	Email  string `json:"email" validate:"required,email"`
	FileID string `json:"file_id" validate:"required"`
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
	UserID string `json:"user_id" validate:"required"`
	Email  string `json:"email" validate:"required"`
	FileID string `json:"file_id" validate:"required"`
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

type (
	GetStorageInfoRequest struct {
		UserID string `json:"user_id"`
		Email  string `json:"email"`
	}

	ListStorageInfoRequest struct {
		UserID string `json:"user_id"`
	}

	StorageInfo struct {
		ID              string `json:"id"`
		Limit           int64  `json:"limit"`
		Usage           int64  `json:"usage"`
		Remaining       int64  `json:"remaining"`
		UsagePercentage int64  `json:"usage_percentage"`
		IsUnlimited     bool   `json:"is_unlimited"`
	}
)

func (service *GoogleDriveService) ListStorageInfo(ctx context.Context, req *ListStorageInfoRequest) ([]StorageInfo, error) {
	type listUserEmailRequest struct {
		Email string `json:"email"`
	}

	listUserEmail := make([]listUserEmailRequest, 0)
	if err := service.DB.
		Model(&OAuthToken{}).
		Where("user_id = ?", req.UserID).
		Select("email").
		Scan(&listUserEmail).
		Error; err != nil {
		return nil, err
	}

	var (
		mu          sync.Mutex
		listStorage []StorageInfo
	)

	g, gCtx := errgroup.WithContext(ctx)

	for _, email := range listUserEmail {
		// Capture email untuk digunakan di goroutine
		email := email.Email

		g.Go(func() error {
			storageInfo, err := service.GetStorageInfo(gCtx, &GetStorageInfoRequest{
				UserID: req.UserID,
				Email:  email,
			})
			if err != nil {
				return err
			}

			mu.Lock()
			listStorage = append(listStorage, *storageInfo)
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return listStorage, nil
}

func (service *GoogleDriveService) GetStorageInfo(ctx context.Context, req *GetStorageInfoRequest) (*StorageInfo, error) {
	newDriveServiceReq := newDriveServiceRequest{
		UserID: req.UserID,
		Email:  req.Email,
	}

	srv, err := service.newDriveService(ctx, &newDriveServiceReq)
	if err != nil {
		return nil, fmt.Errorf("error creating Google Drive service: %w", err)
	}

	var oauthToken OAuthToken

	service.DB.
		Model(&OAuthToken{}).
		Where("user_id = ? AND email = ?", req.UserID, req.Email).
		First(&oauthToken)

	about := srv.About.Get()
	aboutResult, err := about.Fields("storageQuota").Do()
	if err != nil {
		return nil, fmt.Errorf("unable to get About info: %w", err)
	}

	quota := aboutResult.StorageQuota

	info := &StorageInfo{
		ID:    oauthToken.ID,
		Limit: quota.Limit,
		Usage: quota.Usage,
	}

	if quota.Limit == 0 || quota.Limit == -1 {
		info.IsUnlimited = true
	} else {
		info.Remaining = quota.Limit - quota.Usage
		info.UsagePercentage = (quota.Usage * 100) / quota.Limit
	}

	return info, nil
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
