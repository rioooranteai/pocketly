# Pocketly API Contract

Kontrak ini menjelaskan **apa yang dikirim frontend**, **ke mana dikirim**, dan **apa saja yang bisa dikembalikan API**, termasuk setiap kemungkinan error. Semua isi dokumen ini diambil dari perilaku backend saat ini.

- **Versi API:** `v1`
- **Terakhir diperbarui:** 2026-09-25

---

## Daftar Isi

1. [Konvensi Umum](#1-konvensi-umum)
2. [Autentikasi](#2-autentikasi)
3. [Model Data](#3-model-data)
4. [Endpoint Auth](#4-endpoint-auth)
   - [POST /register](#41-post-register)
   - [POST /login](#42-post-login)
5. [Endpoint Transaksi](#5-endpoint-transaksi)
   - [POST /transactions](#51-post-transactions)
   - [GET /transactions](#52-get-transactions)
   - [GET /transactions/:id](#53-get-transactionsid)
   - [PUT /transactions/:id](#54-put-transactionsid)
   - [DELETE /transactions/:id](#55-delete-transactionsid)
   - [POST /transactions/scan](#56-post-transactionsscan)
6. [Ringkasan Status Code](#6-ringkasan-status-code)
7. [Panduan Penanganan di Frontend](#7-panduan-penanganan-di-frontend)

---

## 1. Konvensi Umum

### Base URL

```
http://<host>:<APP_PORT>/api/v1
```

Contoh lokal: `http://localhost:8080/api/v1`. Port mengikuti `APP_PORT` di `.env` backend.

### Format request dan response

| Hal | Aturan |
|---|---|
| Body request | JSON, header `Content-Type: application/json`. Pengecualian: `/transactions/scan` memakai `multipart/form-data`. |
| Body response | JSON (`application/json; charset=utf-8`). Pengecualian: `DELETE` sukses (204) tidak punya body, dan route yang tidak ada (lihat di bawah). |
| Nama field | `snake_case`, contoh `total_amount`. |
| ID | String UUID, contoh `"6f87a151-8d04-4094-a584-14b885efb158"`. |
| Tanggal | String **RFC 3339 lengkap** dengan jam dan zona waktu, contoh `"2026-09-25T10:00:00+07:00"` atau `"2026-09-25T03:00:00Z"`. Tanggal saja (`"2026-09-25"`) **ditolak**. Zona waktu yang dikirim **dipertahankan** di response. |
| Angka uang | Number JSON (boleh desimal). Harus ≥ 0. |
| Field tak dikenal | Diabaikan, bukan error. |

### Bentuk error

Semua error dari API berbentuk JSON dengan satu field `error`:

```json
{ "error": "pesan yang bisa dibaca manusia" }
```

Pengecualian: **route atau method yang tidak ada** (misalnya `PATCH /transactions/:id` atau `/api/v1/salah`) dijawab `404` dengan body **teks biasa** `404 page not found`, bukan JSON. Jangan memanggil `JSON.parse` sebelum mengecek `Content-Type`.

```http
HTTP/1.1 404 Not Found
Content-Type: text/plain

404 page not found
```

### Pesan error validasi

Kalau body tidak lolos validasi, API mengembalikan `400` dengan pesan per field. Jika ada beberapa kesalahan sekaligus, pesannya **digabung dengan `; `**:

```json
{ "error": "description is required; date is required; items is required" }
```

Pola pesan yang mungkin muncul:

| Pola | Contoh |
|---|---|
| `<field> is required` | `price is required` |
| `<field> must be at least <n> characters` | `password must be at least 8 characters` |
| `<field> must be at most <n> characters` | `name must be at most 100 characters` |
| `<field> must be at least <n> items` | `items must be at least 1 items` |
| `<field> must be at least <n>` (untuk angka) | `quantity must be at least 0` |
| `<field> is invalid` | aturan lain yang tidak lolos |
| `invalid request body` | JSON rusak, atau tipe data salah (misalnya `quantity` berupa `1.5` atau `"1"`, atau format tanggal salah) |

`<field>` adalah nama field dalam huruf kecil. Untuk field di dalam `items`, nama field-nya ditulis tanpa indeks, contohnya `price is required`.

### Batas ukuran body

| Endpoint | Batas | Jika terlewati |
|---|---|---|
| `POST /register`, `POST /login` | 4 KB | `413` `{"error":"request body too large"}` |
| `POST /transactions`, `PUT /transactions/:id` | 64 KB | `413` `{"error":"request body too large"}` |
| `POST /transactions/scan` | file gambar 5 MB (5.242.880 byte) | `413` `{"error":"receipt image too large"}` (lihat detail di endpoint scan) |

### Rate limit

| Endpoint | Batas | Dihitung per |
|---|---|---|
| `POST /register` | 5 request / menit | IP |
| `POST /login` | 10 request / menit | IP |
| `POST /transactions/scan` | 10 request / menit | user yang login |

Semua request dihitung, baik yang berhasil maupun yang gagal. Jika batas terlewati:

```
HTTP/1.1 429 Too Many Requests
Retry-After: 42
```
```json
{ "error": "too many requests, please try again later" }
```

`Retry-After` berisi **jumlah detik** sampai boleh mencoba lagi. Header ini bisa dibaca dari browser karena sudah diekspos lewat CORS.

### CORS

Saat ini API menerima request dari **semua origin** (`*`). Pengaturan ini hanya untuk development, dan akan dibatasi ke domain frontend sebelum rilis. Header yang diizinkan: `Authorization` dan `Content-Type`. Method yang diizinkan: `GET`, `POST`, `PUT`, `DELETE`, dan `OPTIONS`.

---

## 2. Autentikasi

Endpoint transaksi membutuhkan token JWT yang didapat dari `/register` atau `/login`.

```
Authorization: Bearer <token>
```

| Hal | Nilai |
|---|---|
| Masa berlaku token | **12 jam** sejak dibuat |
| Refresh token | Tidak ada. Setelah kedaluwarsa, user harus login ulang. |
| Skema | `Bearer`, tidak peka huruf besar/kecil (`bearer` juga diterima) |

Error autentikasi (berlaku untuk **semua** endpoint `/transactions...`):

| Status | `error` | Kapan |
|---|---|---|
| `401` | `missing authorization header` | Header `Authorization` tidak dikirim |
| `401` | `authorization header must be in the format: Bearer <token>` | Skemanya bukan `Bearer` atau token kosong |
| `401` | `invalid or expired token` | Token rusak, tanda tangannya salah, atau sudah lewat 12 jam |

**Contoh respons gagal**

**401: header Authorization tidak dikirim**

```http
HTTP/1.1 401 Unauthorized
Content-Type: application/json; charset=utf-8

{ "error": "missing authorization header" }
```

**401: skema bukan Bearer**

```http
HTTP/1.1 401 Unauthorized
Content-Type: application/json; charset=utf-8

{ "error": "authorization header must be in the format: Bearer <token>" }
```

**401: token rusak atau kedaluwarsa**

```http
HTTP/1.1 401 Unauthorized
Content-Type: application/json; charset=utf-8

{ "error": "invalid or expired token" }
```

---

## 3. Model Data

### AuthResponse

Dikembalikan oleh `/register` dan `/login`. **Tidak dibungkus `data`.**

| Field | Tipe | Keterangan |
|---|---|---|
| `name` | string | Nama user (sudah di-trim) |
| `email` | string | Email dalam huruf kecil dan sudah di-trim |
| `token` | string | JWT untuk header `Authorization` |

### Transaction

| Field | Tipe | Keterangan |
|---|---|---|
| `id` | string (UUID) | ID transaksi |
| `description` | string | Deskripsi (sudah di-trim) |
| `category` | string | Diisi otomatis oleh server, lihat [Kategori](#kategori). Frontend **tidak** mengirim field ini. |
| `total_amount` | number | Dihitung server: jumlah `quantity × price` dari semua item. Frontend **tidak** mengirim field ini. |
| `date` | string (RFC 3339) | Tanggal transaksi, dengan zona waktu seperti yang dikirim |
| `items` | TransactionItem[] | Minimal 1 item |

### TransactionItem

| Field | Tipe | Keterangan |
|---|---|---|
| `name` | string | Nama item (sudah di-trim) |
| `quantity` | integer | ≥ 0 |
| `price` | number | Harga satuan, ≥ 0 |

Item di response **tidak punya `id`**. Item selalu dikirim ulang secara utuh saat update, lihat [PUT](#54-put-transactionsid).

### Kategori

Nilai `category` selalu salah satu dari daftar berikut:

| Nilai | Arti |
|---|---|
| `food` | Makanan & minuman |
| `transportation` | Transportasi |
| `shopping` | Belanja non-makanan |
| `bills` | Tagihan |
| `entertainment` | Hiburan |
| `health` | Kesehatan |
| `uncategorized` | Kategori tidak bisa ditentukan |

Kategori ditentukan dari `description` setiap kali transaksi dibuat atau diubah. Frontend sebaiknya punya tampilan untuk `uncategorized`.

---

## 4. Endpoint Auth

Kedua endpoint ini **publik** (tidak butuh token).

### 4.1 POST /register

Membuat akun baru dan langsung login. Response-nya sudah berisi token, jadi tidak perlu memanggil `/login` lagi.

**Request**

```
POST /api/v1/register
Content-Type: application/json
```

| Field | Tipe | Wajib | Aturan |
|---|---|---|---|
| `name` | string | ✅ | Maks. 100 karakter, tidak boleh hanya berisi spasi |
| `email` | string | ✅ | Maks. 254 karakter, format email valid. Huruf besar/kecil dan spasi di pinggir diabaikan. |
| `password` | string | ✅ | 8–128 karakter |

```json
{
  "name": "Budi",
  "email": "budi@example.com",
  "password": "rahasia123"
}
```

**Response sukses: `201 Created`**

```json
{
  "name": "Budi",
  "email": "budi@example.com",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Kemungkinan error**

| Status | `error` | Kapan |
|---|---|---|
| `400` | pesan validasi, misalnya `password must be at least 8 characters` | Field kosong atau di luar batas panjang |
| `400` | `invalid email format` | Format email salah (dicek setelah validasi field lolos) |
| `400` | `name must not be empty` | Nama hanya berisi spasi |
| `400` | `invalid request body` | JSON rusak atau tipe data salah |
| `409` | `email already registered` | Email sudah dipakai |
| `413` | `request body too large` | Body > 4 KB |
| `429` | `too many requests, please try again later` | > 5 request/menit dari IP yang sama |
| `500` | `something went wrong` | Kesalahan server |

**Contoh respons gagal**

**400: validasi field (beberapa sekaligus)**

```http
HTTP/1.1 400 Bad Request
Content-Type: application/json; charset=utf-8

{ "error": "name is required; password must be at least 8 characters" }
```

**400: format email salah**

```http
HTTP/1.1 400 Bad Request
Content-Type: application/json; charset=utf-8

{ "error": "invalid email format" }
```

**400: nama hanya berisi spasi**

```http
HTTP/1.1 400 Bad Request
Content-Type: application/json; charset=utf-8

{ "error": "name must not be empty" }
```

**400: JSON rusak atau tipe data salah**

```http
HTTP/1.1 400 Bad Request
Content-Type: application/json; charset=utf-8

{ "error": "invalid request body" }
```

**409: email sudah dipakai**

```http
HTTP/1.1 409 Conflict
Content-Type: application/json; charset=utf-8

{ "error": "email already registered" }
```

**413: body lebih dari 4 KB**

```http
HTTP/1.1 413 Request Entity Too Large
Content-Type: application/json; charset=utf-8

{ "error": "request body too large" }
```

**429: rate limit**

```http
HTTP/1.1 429 Too Many Requests
Content-Type: application/json; charset=utf-8
Retry-After: 42

{ "error": "too many requests, please try again later" }
```

**500: kesalahan server**

```http
HTTP/1.1 500 Internal Server Error
Content-Type: application/json; charset=utf-8

{ "error": "something went wrong" }
```

### 4.2 POST /login

**Request**

```
POST /api/v1/login
Content-Type: application/json
```

| Field | Tipe | Wajib | Aturan |
|---|---|---|---|
| `email` | string | ✅ | Maks. 254 karakter. Huruf besar/kecil dan spasi di pinggir diabaikan. |
| `password` | string | ✅ | Maks. 128 karakter |

```json
{
  "email": "budi@example.com",
  "password": "rahasia123"
}
```

**Response sukses: `200 OK`**

```json
{
  "name": "Budi",
  "email": "budi@example.com",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Kemungkinan error**

| Status | `error` | Kapan |
|---|---|---|
| `400` | pesan validasi, misalnya `email is required` | Field kosong atau terlalu panjang |
| `400` | `invalid request body` | JSON rusak atau tipe data salah |
| `401` | `invalid email or password` | Email tidak terdaftar **atau** password salah. Sengaja tidak dibedakan. |
| `413` | `request body too large` | Body > 4 KB |
| `429` | `too many requests, please try again later` | > 10 request/menit dari IP yang sama |
| `500` | `something went wrong` | Kesalahan server |

**Contoh respons gagal**

**400: validasi field**

```http
HTTP/1.1 400 Bad Request
Content-Type: application/json; charset=utf-8

{ "error": "email is required" }
```

**400: JSON rusak atau tipe data salah**

```http
HTTP/1.1 400 Bad Request
Content-Type: application/json; charset=utf-8

{ "error": "invalid request body" }
```

**401: email tidak terdaftar atau password salah**

```http
HTTP/1.1 401 Unauthorized
Content-Type: application/json; charset=utf-8

{ "error": "invalid email or password" }
```

**413: body lebih dari 4 KB**

```http
HTTP/1.1 413 Request Entity Too Large
Content-Type: application/json; charset=utf-8

{ "error": "request body too large" }
```

**429: rate limit**

```http
HTTP/1.1 429 Too Many Requests
Content-Type: application/json; charset=utf-8
Retry-After: 42

{ "error": "too many requests, please try again later" }
```

**500: kesalahan server**

```http
HTTP/1.1 500 Internal Server Error
Content-Type: application/json; charset=utf-8

{ "error": "something went wrong" }
```

---

## 5. Endpoint Transaksi

Semua endpoint di bagian ini **wajib** memakai header `Authorization: Bearer <token>`, dan bisa mengembalikan error `401` dari [Autentikasi](#2-autentikasi). User hanya bisa melihat dan mengubah transaksinya sendiri.

### 5.1 POST /transactions

Mencatat transaksi yang diisi manual.

**Request**

```
POST /api/v1/transactions
Authorization: Bearer <token>
Content-Type: application/json
```

| Field | Tipe | Wajib | Aturan |
|---|---|---|---|
| `description` | string | ✅ | Tidak boleh kosong atau hanya berisi spasi. Dipakai untuk menentukan kategori. |
| `date` | string | ✅ | RFC 3339 lengkap, contoh `"2026-09-25T12:30:00+07:00"` |
| `items` | array | ✅ | Minimal 1 item |
| `items[].name` | string | ✅ | Tidak boleh kosong atau hanya berisi spasi |
| `items[].quantity` | integer | ✅ | ≥ 0, harus bilangan bulat |
| `items[].price` | number | ✅ | ≥ 0. Nilai `0` boleh, misalnya untuk item gratis. |

```json
{
  "description": "Makan siang di warung",
  "date": "2026-09-25T12:30:00+07:00",
  "items": [
    { "name": "Nasi goreng", "quantity": 2, "price": 15000 },
    { "name": "Es teh", "quantity": 1, "price": 0 }
  ]
}
```

**Response sukses: `201 Created`**

```json
{
  "message": "Transaction created successfully",
  "data": {
    "id": "6f87a151-8d04-4094-a584-14b885efb158",
    "description": "Makan siang di warung",
    "category": "food",
    "total_amount": 30000,
    "date": "2026-09-25T12:30:00+07:00",
    "items": [
      { "name": "Nasi goreng", "quantity": 2, "price": 15000 },
      { "name": "Es teh", "quantity": 1, "price": 0 }
    ]
  }
}
```

**Kemungkinan error**

| Status | `error` | Kapan |
|---|---|---|
| `400` | pesan validasi, misalnya `price is required; quantity must be at least 0` | Field wajib tidak ada, atau angka negatif |
| `400` | `invalid request body` | JSON rusak, `quantity` bukan bilangan bulat, atau format tanggal salah |
| `400` | `description must not be empty` | Deskripsi hanya berisi spasi |
| `400` | `item <n>: item needs a name and a non-negative quantity and price within range` | Item ke-`n` (mulai dari 1) namanya hanya berisi spasi, atau harganya terlalu besar |
| `400` | `transaction total is too large` | Total semua item terlalu besar untuk disimpan |
| `413` | `request body too large` | Body > 64 KB |
| `500` | `Failed to create transaction` | Kesalahan server |

**Contoh respons gagal**

**400: body kosong `{}`**

```http
HTTP/1.1 400 Bad Request
Content-Type: application/json; charset=utf-8

{ "error": "description is required; date is required; items is required" }
```

**400: field item hilang atau negatif**

```http
HTTP/1.1 400 Bad Request
Content-Type: application/json; charset=utf-8

{ "error": "price is required; quantity must be at least 0" }
```

**400: contoh: `"date": "2026-09-25"` atau `"quantity": 1.5`**

```http
HTTP/1.1 400 Bad Request
Content-Type: application/json; charset=utf-8

{ "error": "invalid request body" }
```

**400: deskripsi hanya berisi spasi**

```http
HTTP/1.1 400 Bad Request
Content-Type: application/json; charset=utf-8

{ "error": "description must not be empty" }
```

**400: item ke-2 tidak valid**

```http
HTTP/1.1 400 Bad Request
Content-Type: application/json; charset=utf-8

{ "error": "item 2: item needs a name and a non-negative quantity and price within range" }
```

**400: total terlalu besar**

```http
HTTP/1.1 400 Bad Request
Content-Type: application/json; charset=utf-8

{ "error": "transaction total is too large" }
```

**413: body lebih dari 64 KB**

```http
HTTP/1.1 413 Request Entity Too Large
Content-Type: application/json; charset=utf-8

{ "error": "request body too large" }
```

**500: kesalahan server**

```http
HTTP/1.1 500 Internal Server Error
Content-Type: application/json; charset=utf-8

{ "error": "Failed to create transaction" }
```

### 5.2 GET /transactions

Mengambil semua transaksi milik user yang sedang login, **diurutkan dari tanggal terbaru**. Endpoint ini belum punya paginasi.

**Request**

```
GET /api/v1/transactions
Authorization: Bearer <token>
```

**Response sukses: `200 OK`**

```json
{
  "data": [
    {
      "id": "6f87a151-8d04-4094-a584-14b885efb158",
      "description": "Makan siang di warung",
      "category": "food",
      "total_amount": 30000,
      "date": "2026-09-25T12:30:00+07:00",
      "items": [
        { "name": "Nasi goreng", "quantity": 2, "price": 15000 }
      ]
    }
  ]
}
```

Jika belum ada transaksi, `data` berisi array kosong `[]`, bukan `null`.

**Kemungkinan error**

| Status | `error` | Kapan |
|---|---|---|
| `500` | `Failed to fetch transactions` | Kesalahan server |

**Contoh respons gagal**

**500: kesalahan server**

```http
HTTP/1.1 500 Internal Server Error
Content-Type: application/json; charset=utf-8

{ "error": "Failed to fetch transactions" }
```

### 5.3 GET /transactions/:id

**Request**

```
GET /api/v1/transactions/6f87a151-8d04-4094-a584-14b885efb158
Authorization: Bearer <token>
```

**Response sukses: `200 OK`**

```json
{
  "data": {
    "id": "6f87a151-8d04-4094-a584-14b885efb158",
    "description": "Makan siang di warung",
    "category": "food",
    "total_amount": 30000,
    "date": "2026-09-25T12:30:00+07:00",
    "items": [
      { "name": "Nasi goreng", "quantity": 2, "price": 15000 }
    ]
  }
}
```

**Kemungkinan error**

| Status | `error` | Kapan |
|---|---|---|
| `404` | `Transaction not found` | ID tidak ada, **atau** transaksinya milik user lain. Keduanya sengaja dijawab sama. |
| `500` | `Failed to fetch transaction` | Kesalahan server |

**Contoh respons gagal**

**404: ID tidak ada atau milik user lain**

```http
HTTP/1.1 404 Not Found
Content-Type: application/json; charset=utf-8

{ "error": "Transaction not found" }
```

**500: kesalahan server**

```http
HTTP/1.1 500 Internal Server Error
Content-Type: application/json; charset=utf-8

{ "error": "Failed to fetch transaction" }
```

### 5.4 PUT /transactions/:id

Mengganti isi transaksi. Body-nya **sama persis** dengan [POST /transactions](#51-post-transactions), dan semua field wajib dikirim.

Hal yang perlu diperhatikan:
- **Semua item diganti.** Item lama dihapus dan diganti dengan daftar `items` yang dikirim. Untuk mengubah satu item saja, kirim ulang seluruh daftar item.
- **Kategori dihitung ulang** dari `description` yang baru.
- **`total_amount` dihitung ulang** dari item yang baru.

**Request**

```
PUT /api/v1/transactions/6f87a151-8d04-4094-a584-14b885efb158
Authorization: Bearer <token>
Content-Type: application/json
```

```json
{
  "description": "Bensin Pertamina",
  "date": "2026-09-25T08:00:00+07:00",
  "items": [
    { "name": "Pertalite", "quantity": 1, "price": 50000 }
  ]
}
```

**Response sukses: `200 OK`**

```json
{
  "message": "Transaction updated successfully",
  "data": {
    "id": "6f87a151-8d04-4094-a584-14b885efb158",
    "description": "Bensin Pertamina",
    "category": "transportation",
    "total_amount": 50000,
    "date": "2026-09-25T08:00:00+07:00",
    "items": [
      { "name": "Pertalite", "quantity": 1, "price": 50000 }
    ]
  }
}
```

**Kemungkinan error**

Semua error `400` dan `413` dari [POST /transactions](#51-post-transactions) juga berlaku di sini, ditambah:

| Status | `error` | Kapan |
|---|---|---|
| `404` | `Transaction not found` | ID tidak ada, atau transaksinya milik user lain |
| `500` | `Failed to update transaction` | Kesalahan server |

**Contoh respons gagal**

**400: item tidak valid (semua contoh 400 dari POST juga berlaku)**

```http
HTTP/1.1 400 Bad Request
Content-Type: application/json; charset=utf-8

{ "error": "item 1: item needs a name and a non-negative quantity and price within range" }
```

**404: ID tidak ada atau milik user lain**

```http
HTTP/1.1 404 Not Found
Content-Type: application/json; charset=utf-8

{ "error": "Transaction not found" }
```

**413: body lebih dari 64 KB**

```http
HTTP/1.1 413 Request Entity Too Large
Content-Type: application/json; charset=utf-8

{ "error": "request body too large" }
```

**500: kesalahan server**

```http
HTTP/1.1 500 Internal Server Error
Content-Type: application/json; charset=utf-8

{ "error": "Failed to update transaction" }
```

### 5.5 DELETE /transactions/:id

**Request**

```
DELETE /api/v1/transactions/6f87a151-8d04-4094-a584-14b885efb158
Authorization: Bearer <token>
```

**Response sukses: `204 No Content`**, tanpa body.

**Kemungkinan error**

| Status | `error` | Kapan |
|---|---|---|
| `404` | `Transaction not found` | ID tidak ada, atau transaksinya milik user lain |
| `500` | `Failed to delete transaction` | Kesalahan server |

**Contoh respons gagal**

**404: ID tidak ada atau milik user lain**

```http
HTTP/1.1 404 Not Found
Content-Type: application/json; charset=utf-8

{ "error": "Transaction not found" }
```

**500: kesalahan server**

```http
HTTP/1.1 500 Internal Server Error
Content-Type: application/json; charset=utf-8

{ "error": "Failed to delete transaction" }
```

### 5.6 POST /transactions/scan

Membuat transaksi dari **foto struk**. Server membaca deskripsi dan daftar item dari gambar memakai AI, lalu menyimpannya seperti transaksi biasa.

Hal yang perlu diperhatikan:
- **Tanggal transaksi = waktu scan** (UTC). Struk tidak dibaca tanggalnya.
- **Proses ini bisa lama, sampai sekitar 25 detik.** Tampilkan indikator loading, dan set timeout di sisi frontend minimal **30 detik**.
- **Jika struk gagal dibaca, tidak ada yang disimpan.** API mengembalikan `422`, dan frontend bisa meminta user memotret ulang atau mengisi manual lewat [POST /transactions](#51-post-transactions).
- Dibatasi **10 scan per menit per user**. Percobaan yang gagal juga ikut dihitung.

**Request**

```
POST /api/v1/transactions/scan
Authorization: Bearer <token>
Content-Type: multipart/form-data
```

| Field (form) | Tipe | Wajib | Aturan |
|---|---|---|---|
| `receipt` | file | ✅ | Gambar struk, maks. **5 MB**. Format yang disarankan: JPEG, PNG, atau WebP. |

Contoh dengan `fetch`:

```js
const form = new FormData();
form.append("receipt", fileInput.files[0]);

const res = await fetch(`${BASE_URL}/transactions/scan`, {
  method: "POST",
  headers: { Authorization: `Bearer ${token}` }, // jangan set Content-Type manual
  body: form,
});
```

Jangan mengisi header `Content-Type` secara manual. Browser akan mengisinya sendiri beserta `boundary` multipart.

**Response sukses: `201 Created`**

```json
{
  "message": "Transaction created successfully from receipt",
  "data": {
    "id": "99f91ac9-6088-46e0-809f-d8cee088d342",
    "description": "Indomaret",
    "category": "food",
    "total_amount": 45000,
    "date": "2026-09-25T05:12:44Z",
    "items": [
      { "name": "Beras 5kg", "quantity": 1, "price": 40000 },
      { "name": "Minyak goreng", "quantity": 1, "price": 5000 }
    ]
  }
}
```

**Kemungkinan error**

| Status | `error` | Kapan |
|---|---|---|
| `400` | `Missing or invalid file field 'receipt'` | Field `receipt` tidak ada, atau request bukan multipart |
| `400` | `image data is empty` | File kosong (0 byte) |
| `400` | `image size exceeds maximum allowed limit` | File sedikit di atas 5 MB (sampai 5 MB + 64 KB) |
| `413` | `receipt image too large` | File jauh di atas 5 MB. Upload dihentikan sebelum selesai dibaca. |
| `422` | `receipt could not be read as a valid transaction: transaction must have at least one item` | Tidak ada item yang terbaca dari struk |
| `422` | `receipt could not be read as a valid transaction: description must not be empty` | Deskripsi struk tidak terbaca |
| `422` | `receipt could not be read as a valid transaction: item <n>: ...` | Item ke-`n` hasil bacaan tidak valid (nama kosong atau harga negatif) |
| `422` | `receipt could not be read as a valid transaction: transaction total is too large` | Total hasil bacaan terlalu besar |
| `429` | `too many requests, please try again later` | > 10 scan/menit untuk user ini (lihat `Retry-After`) |
| `500` | `Failed to process receipt image` | Layanan AI gagal atau melewati batas waktu 25 detik, atau file bukan gambar yang didukung |
| `500` | `Failed to open uploaded file` / `Failed to read uploaded file` | Kesalahan server saat membaca file |

**Contoh respons gagal**

**400: field receipt tidak ada**

```http
HTTP/1.1 400 Bad Request
Content-Type: application/json; charset=utf-8

{ "error": "Missing or invalid file field 'receipt'" }
```

**400: file 0 byte**

```http
HTTP/1.1 400 Bad Request
Content-Type: application/json; charset=utf-8

{ "error": "image data is empty" }
```

**400: file sedikit di atas 5 MB**

```http
HTTP/1.1 400 Bad Request
Content-Type: application/json; charset=utf-8

{ "error": "image size exceeds maximum allowed limit" }
```

**413: file jauh di atas 5 MB**

```http
HTTP/1.1 413 Request Entity Too Large
Content-Type: application/json; charset=utf-8

{ "error": "receipt image too large" }
```

**422: tidak ada item terbaca**

```http
HTTP/1.1 422 Unprocessable Entity
Content-Type: application/json; charset=utf-8

{ "error": "receipt could not be read as a valid transaction: transaction must have at least one item" }
```

**422: deskripsi tidak terbaca**

```http
HTTP/1.1 422 Unprocessable Entity
Content-Type: application/json; charset=utf-8

{ "error": "receipt could not be read as a valid transaction: description must not be empty" }
```

**422: item hasil bacaan tidak valid**

```http
HTTP/1.1 422 Unprocessable Entity
Content-Type: application/json; charset=utf-8

{ "error": "receipt could not be read as a valid transaction: item 1: item needs a name and a non-negative quantity and price within range" }
```

**422: total hasil bacaan terlalu besar**

```http
HTTP/1.1 422 Unprocessable Entity
Content-Type: application/json; charset=utf-8

{ "error": "receipt could not be read as a valid transaction: transaction total is too large" }
```

**429: lebih dari 10 scan/menit**

```http
HTTP/1.1 429 Too Many Requests
Content-Type: application/json; charset=utf-8
Retry-After: 42

{ "error": "too many requests, please try again later" }
```

**500: layanan AI gagal atau timeout**

```http
HTTP/1.1 500 Internal Server Error
Content-Type: application/json; charset=utf-8

{ "error": "Failed to process receipt image" }
```

**500: kesalahan server saat membaca file**

```http
HTTP/1.1 500 Internal Server Error
Content-Type: application/json; charset=utf-8

{ "error": "Failed to read uploaded file" }
```

Untuk memudahkan penanganan, semua `422` diawali teks yang sama: `receipt could not be read as a valid transaction:`.

---

## 6. Ringkasan Status Code

| Status | Arti | Body |
|---|---|---|
| `200` | Berhasil (GET, PUT, login) | JSON |
| `201` | Berhasil membuat data (register, create, scan) | JSON |
| `204` | Berhasil menghapus | Kosong |
| `400` | Input tidak valid | `{"error": ...}` |
| `401` | Token tidak ada, salah, kedaluwarsa, atau login gagal | `{"error": ...}` |
| `404` | Transaksi tidak ditemukan atau bukan milik user | `{"error": ...}` |
| `404` | Route atau method tidak ada | Teks `404 page not found` |
| `409` | Email sudah terdaftar | `{"error": ...}` |
| `413` | Body atau file terlalu besar | `{"error": ...}` |
| `422` | Struk tidak bisa dibaca menjadi transaksi yang valid | `{"error": ...}` |
| `429` | Terlalu banyak request, lihat header `Retry-After` | `{"error": ...}` |
| `500` | Kesalahan server | `{"error": ...}` |

---

## 7. Panduan Penanganan di Frontend

- **`401` di endpoint transaksi:** hapus token yang tersimpan dan arahkan user ke halaman login. Token berlaku 12 jam, dan belum ada refresh token.
- **`401` di `/login`:** tampilkan "Email atau password salah". API sengaja tidak memberi tahu mana yang salah.
- **`400`:** pesan di `error` sudah bisa ditampilkan langsung ke user, atau dipetakan ke field form. Kalau ingin menandai field tertentu, pecah pesannya dengan `"; "`.
- **`404` untuk transaksi:** anggap transaksinya sudah tidak ada, lalu keluarkan dari tampilan. Misalnya, transaksinya sudah dihapus di tab lain.
- **`413` dan batas 5 MB:** cek ukuran file di frontend sebelum upload, supaya user tidak perlu menunggu upload yang pasti ditolak. Kompres gambar kalau perlu.
- **`422` di `/scan`:** tawarkan dua pilihan ke user, yaitu "Foto ulang" atau "Isi manual".
- **`429`:** nonaktifkan tombol terkait selama `Retry-After` detik, dan tampilkan hitung mundur.
- **`500`:** tampilkan pesan umum dan tombol "Coba lagi".
- **Jangan kirim** `category`, `total_amount`, `id`, atau `user_id` di body. Field tersebut dihitung server, dan kalau dikirim akan diabaikan.
- **Nilai yang di-trim:** `description`, nama item, dan nama user di-trim oleh server, dan email diubah ke huruf kecil. Selalu tampilkan nilai dari response, bukan dari input form.
