# Deployment Guide - Sistem Akuntansi

Panduan lengkap step-by-step untuk deploy aplikasi Sistem Akuntansi menggunakan Docker di VPS.

---

## Step 1: Persiapan VPS

### 1.1 Login ke VPS

```bash
ssh root@YOUR_VPS_IP
# atau
ssh username@YOUR_VPS_IP
```

### 1.2 Update System

```bash
# Ubuntu/Debian
sudo apt update && sudo apt upgrade -y

# CentOS/RHEL
sudo yum update -y
```

---

## Step 2: Install Docker

### Ubuntu/Debian

```bash
# Install dependencies
sudo apt install -y apt-transport-https ca-certificates curl software-properties-common

# Add Docker GPG key
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg

# Add Docker repository
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# Install Docker
sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io

# Start Docker
sudo systemctl start docker
sudo systemctl enable docker

# Verify installation
docker --version
```

### CentOS/RHEL

```bash
# Install dependencies
sudo yum install -y yum-utils

# Add Docker repository
sudo yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo

# Install Docker
sudo yum install -y docker-ce docker-ce-cli containerd.io

# Start Docker
sudo systemctl start docker
sudo systemctl enable docker

# Verify installation
docker --version
```

### (Optional) Run Docker tanpa sudo

```bash
sudo usermod -aG docker $USER
# Logout dan login kembali agar perubahan berlaku
exit
# Login lagi
ssh username@YOUR_VPS_IP
```

---

## Step 3: Install Docker Compose

```bash
# Download Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose

# Beri permission executable
sudo chmod +x /usr/local/bin/docker-compose

# Verify installation
docker-compose --version
```

---

## Step 4: Install Git

```bash
# Ubuntu/Debian
sudo apt install -y git

# CentOS/RHEL
sudo yum install -y git

# Verify installation
git --version
```

---

## Step 5: Clone Repository

```bash
# Buat direktori untuk aplikasi
mkdir -p /var/www
cd /var/www

# Clone repository
git clone https://github.com/Jerly08/sistem_akuntansi.git

# Masuk ke direktori
cd sistem_akuntansi
```

---

## Step 6: Setup Environment Variables

### 6.1 Copy template environment

```bash
cp .env.example .env
```

### 6.2 Generate JWT Secrets

```bash
# Generate random secrets
echo "JWT_SECRET: $(openssl rand -hex 32)"
echo "JWT_ACCESS_SECRET: $(openssl rand -hex 32)"
echo "JWT_REFRESH_SECRET: $(openssl rand -hex 32)"
```

### 6.3 Edit file .env

```bash
nano .env
```

Ubah nilai-nilai berikut:

```env
# Database - GANTI PASSWORD!
DB_USER=postgres
DB_PASSWORD=GantiDenganPasswordKuat123!
DB_NAME=sistem_akuntansi
DB_PORT=5432

# Backend
BACKEND_PORT=8080
ENVIRONMENT=production

# JWT Secrets - PASTE hasil generate di atas
JWT_SECRET=paste_hasil_generate_disini
JWT_ACCESS_SECRET=paste_hasil_generate_disini
JWT_REFRESH_SECRET=paste_hasil_generate_disini

# Frontend
FRONTEND_PORT=3000
NODE_ENV=production

# API URL - GANTI dengan IP VPS Anda
NEXT_PUBLIC_API_URL=http://YOUR_VPS_IP:8080

# CORS - GANTI dengan IP VPS Anda
ALLOWED_ORIGINS=http://YOUR_VPS_IP:3000
```

Simpan: `Ctrl+O`, Enter, `Ctrl+X`

---

## Step 7: Build & Run Docker

### 7.1 Build semua images

```bash
docker-compose build
```

Proses ini akan memakan waktu 5-15 menit tergantung kecepatan VPS.

### 7.2 Jalankan semua services

```bash
docker-compose up -d
```

### 7.3 Cek status containers

```bash
docker-compose ps
```

Output yang diharapkan:
```
NAME                  STATUS              PORTS
akuntansi_db          Up (healthy)        0.0.0.0:5432->5432/tcp
akuntansi_backend     Up                  0.0.0.0:8080->8080/tcp
akuntansi_frontend    Up                  0.0.0.0:3000->3000/tcp
```

### 7.4 Cek logs jika ada error

```bash
# Semua logs
docker-compose logs

# Logs backend saja
docker-compose logs backend

# Logs frontend saja
docker-compose logs frontend

# Follow logs (real-time)
docker-compose logs -f
```

---

## Step 8: Konfigurasi Firewall

### UFW (Ubuntu)

```bash
# Enable UFW
sudo ufw enable

# Allow SSH (PENTING! Jangan sampai terkunci)
sudo ufw allow 22

# Allow aplikasi
sudo ufw allow 3000  # Frontend
sudo ufw allow 8080  # Backend API

# Cek status
sudo ufw status
```

### Firewalld (CentOS)

```bash
sudo firewall-cmd --permanent --add-port=3000/tcp
sudo firewall-cmd --permanent --add-port=8080/tcp
sudo firewall-cmd --reload
```

---

## Step 9: Akses Aplikasi

Buka browser dan akses:

- **Frontend**: `http://YOUR_VPS_IP:3000`
- **Backend API**: `http://YOUR_VPS_IP:8080`
- **API Documentation**: `http://YOUR_VPS_IP:8080/swagger/index.html`

---

## Step 10: (Optional) Setup Domain & SSL

Jika Anda punya domain, ikuti langkah berikut:

### 10.1 Install Nginx

```bash
# Ubuntu/Debian
sudo apt install -y nginx

# Start Nginx
sudo systemctl start nginx
sudo systemctl enable nginx
```

### 10.2 Install Certbot untuk SSL

```bash
# Ubuntu/Debian
sudo apt install -y certbot python3-certbot-nginx
```

### 10.3 Konfigurasi Nginx

```bash
sudo nano /etc/nginx/sites-available/akuntansi
```

Paste konfigurasi berikut:

```nginx
server {
    listen 80;
    server_name your-domain.com www.your-domain.com;

    # Frontend
    location / {
        proxy_pass http://localhost:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }

    # Backend API
    location /api {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Swagger docs
    location /swagger {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
    }
}
```

### 10.4 Enable site & restart Nginx

```bash
sudo ln -s /etc/nginx/sites-available/akuntansi /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx
```

### 10.5 Generate SSL Certificate

```bash
sudo certbot --nginx -d your-domain.com -d www.your-domain.com
```

### 10.6 Update .env untuk HTTPS

```bash
nano .env
```

Ubah:
```env
NEXT_PUBLIC_API_URL=https://your-domain.com
ALLOWED_ORIGINS=https://your-domain.com,https://www.your-domain.com
```

### 10.7 Restart Docker

```bash
docker-compose down
docker-compose up -d
```

---

## Maintenance Commands

### Update Aplikasi

```bash
cd /var/www/sistem_akuntansi

# Pull latest code
git pull origin main

# Rebuild dan restart
docker-compose down
docker-compose up -d --build
```

### Backup Database

```bash
# Backup
docker-compose exec db pg_dump -U postgres sistem_akuntansi > backup_$(date +%Y%m%d).sql

# Restore
docker-compose exec -T db psql -U postgres sistem_akuntansi < backup_20250120.sql
```

### Restart Services

```bash
# Restart semua
docker-compose restart

# Restart satu service
docker-compose restart backend
docker-compose restart frontend
```

### Stop Aplikasi

```bash
docker-compose down
```

### Lihat Resource Usage

```bash
docker stats
```

### Masuk ke Container

```bash
# Database
docker-compose exec db psql -U postgres -d sistem_akuntansi

# Backend
docker-compose exec backend sh

# Frontend
docker-compose exec frontend sh
```

---

## Troubleshooting

### Error: "Cannot connect to database"

```bash
# Cek apakah database sudah ready
docker-compose logs db

# Tunggu sampai "database system is ready to accept connections"
# Lalu restart backend
docker-compose restart backend
```

### Error: "CORS error" di browser

Pastikan `ALLOWED_ORIGINS` di `.env` sudah benar:
```env
ALLOWED_ORIGINS=http://YOUR_VPS_IP:3000
```

### Error: "Connection refused" ke API

1. Cek backend running: `docker-compose ps`
2. Cek logs: `docker-compose logs backend`
3. Pastikan firewall allow port 8080

### Port sudah digunakan

```bash
# Cek port yang digunakan
sudo netstat -tlnp | grep :3000
sudo netstat -tlnp | grep :8080

# Ubah port di .env jika perlu
FRONTEND_PORT=3001
BACKEND_PORT=8081
```

### Container terus restart

```bash
# Lihat logs untuk error
docker-compose logs --tail=100 backend
docker-compose logs --tail=100 frontend
```

---

## Quick Reference

| Command | Deskripsi |
|---------|-----------|
| `docker-compose up -d` | Start semua services |
| `docker-compose down` | Stop semua services |
| `docker-compose ps` | Lihat status containers |
| `docker-compose logs -f` | Lihat logs real-time |
| `docker-compose restart` | Restart semua services |
| `docker-compose build` | Build ulang images |
| `docker-compose up -d --build` | Build & start |

---

## Support

Jika ada masalah, cek:
1. Logs: `docker-compose logs`
2. Status: `docker-compose ps`
3. Resource: `docker stats`
