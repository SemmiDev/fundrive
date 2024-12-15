# Fun Drive

Library Go untuk mengelola integrasi Google Drive dengan banyak akun pengguna.

### Gambaran Umum
Fun Drive menyederhanakan integrasi Google Drive untuk aplikasi yang perlu menangani beberapa akun pengguna dan akses Google Drive mereka.

### Instalasi
```bash
  go get github.com/SemmiDev/fundrive
```

### Latar Belakang

Aplikasi X memerlukan fitur bagi penggunanya untuk dapat mengakses dan mengelola file Google Drive mereka secara langsung melalui aplikasi. Karea pengguna dapat menghubungkan beberapa akun email, mereka mungkin memiliki beberapa akun Google Drive. Library ini membantu Aplikasi X mengelola berbagai aspek termasuk:

- Manajemen token
- Penanganan layanan Google Drive
- Dukungan multi-akun
- Alur autentikasi

### Prasyarat
Sebelum menggunakan library ini, pastikan Anda telah:

- Membuat Project Google Cloud dengan API Google Drive yang sudah diaktifkan.
- Mengatur OAuth consent screen dengan status "Production" pada pengaturan Publishing untuk mencegah kedaluwarsa refresh token.
- Menghasilkan kredensial yang diperlukan (client ID dan client secret)
- Pastikan di client setting AuthCodeOption nya pakai oauth2.AccessTypeOffline untuk obtain the refresh token  dan oauth2.ApprovalForce, for forces the users to view the consent dialog

### Kebutuhan Database
Library ini secara otomatis mengelola tabel `fundrive_oauth_tokens` di database Anda untuk penyimpanan dan pengelolaan token.

### Alur Autentikasi

- Proses login OAuth ditangani oleh aplikasi yang mengimplementasikan (Aplikasi X), dan pastikan telah memenuhi scopes yang diperlukan. Lihat [oauth_config.go](./oauth_config.go)
- Manajemen token pasca-autentikasi dan penanganan layanan Google Drive dikelola oleh library ini.

### Inisialisasi Service
Lihat contoh implementasi di [main.go](./example/main.go)

### Referensi
- [Access token and refresh token explained](https://medium.com/starthinker/google-oauth-2-0-access-token-and-refresh-token-explained-cccf2fc0a6d9)
- [Google api refresh token expiring](https://stackoverflow.com/questions/71375097/google-api-refresh-token-expiring)
- [StorageQuota.limit](https://developers.google.com/drive/api/reference/rest/v3/about)
- [Expiration of authorization code](https://help.developer.intuit.com/s/question/0D50f00004wLKKVCA4/expiration-of-authorization-code)
