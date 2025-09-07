# OTP Authentication Service

یک سرویس احراز هویت مبتنی بر OTP با استفاده از Golang، Gin، PostgreSQL و Redis.

## ویژگی‌ها

- ✅ **OTP Login & Registration**: درخواست و تأیید OTP برای ورود/ثبت‌نام
- ✅ **Rate Limiting**: محدودیت 3 درخواست OTP در 10 دقیقه برای هر شماره
- ✅ **JWT Authentication**: احراز هویت با JWT token
- ✅ **User Management**: مدیریت کاربران با pagination و جستجو
- ✅ **PostgreSQL**: پایگاه‌داده PostgreSQL با GORM
- ✅ **Redis**: ذخیره OTP و rate limiting
- ✅ **Docker**: کانتینری‌سازی با Docker Compose
- ✅ **Swagger**: مستندات API با Swagger

## ساختار پروژه

```
├── cmd/server/           # نقطه ورود اصلی
├── internal/
│   ├── api/             # HTTP handlers و routes
│   ├── middleware/      # JWT middleware
│   ├── models/          # مدل‌های دیتابیس
│   ├── repos/           # repository layer
│   └── services/        # business logic
├── docs/                # مستندات Swagger
├── docker-compose.yml   # تنظیمات Docker
└── Dockerfile          # تصویر Docker
```

## اجرای پروژه

### با Docker Compose (پیشنهادی)

```bash
# کلون کردن پروژه
git clone <repository-url>
cd OTP-Authentication-Service

# اجرای سرویس‌ها
docker compose up --build

# اجرا در background
docker compose up --build -d
```

### اجرای محلی

```bash
# نصب وابستگی‌ها
go mod download

# اجرای PostgreSQL و Redis
docker compose up db redis -d

# تنظیم متغیرهای محیطی
export DATABASE_DSN="host=localhost user=user password=password dbname=otp_service port=5432 sslmode=disable"
export REDIS_ADDR="localhost:6379"
export REDIS_PASS="redispass"
export JWT_SECRET="change_this_to_a_strong_secret"

# اجرای برنامه
go run ./cmd/server
```

## API Endpoints

### احراز هویت

#### درخواست OTP
```bash
POST /auth/request-otp
Content-Type: application/json

{
  "phone": "+14165551234"
}
```

**پاسخ:**
```json
{
  "message": "OTP generated. Check console logs."
}
```

#### تأیید OTP
```bash
POST /auth/verify-otp
Content-Type: application/json

{
  "phone": "+14165551234",
  "otp": "123456"
}
```

**پاسخ:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### مدیریت کاربران (نیاز به JWT)

#### دریافت کاربر
```bash
GET /users/1
Authorization: Bearer <token>
```

#### لیست کاربران
```bash
GET /users?page=1&limit=10&search=+1416
Authorization: Bearer <token>
```

**پاسخ:**
```json
{
  "data": [
    {
      "id": 1,
      "phone": "+14165551234",
      "registered_at": "2025-01-07T12:34:56Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 1
  }
}
```

### Health Check

```bash
GET /health
```

## تنظیمات

### متغیرهای محیطی

| متغیر | پیش‌فرض | توضیح |
|-------|---------|-------|
| `DATABASE_DSN` | - | رشته اتصال PostgreSQL |
| `REDIS_ADDR` | - | آدرس Redis |
| `REDIS_PASS` | - | رمز عبور Redis |
| `JWT_SECRET` | - | کلید مخفی JWT |

### Docker Compose Services

- **app**: سرویس اصلی (پورت 8080)
- **db**: PostgreSQL (پورت 5432)
- **redis**: Redis (پورت 6379)

## تست کردن

```bash
# تست کامل
curl -X POST http://localhost:8080/auth/request-otp \
  -H 'Content-Type: application/json' \
  -d '{"phone":"+14165551234"}'

# بررسی لاگ‌ها برای OTP
# سپس تأیید OTP
curl -X POST http://localhost:8080/auth/verify-otp \
  -H 'Content-Type: application/json' \
  -d '{"phone":"+14165551234","otp":"123456"}'

# استفاده از token برای API های محافظت شده
curl -X GET http://localhost:8080/users \
  -H 'Authorization: Bearer <token>'
```

## مستندات Swagger

بعد از اجرای سرویس، مستندات Swagger در آدرس زیر در دسترس است:

```
http://localhost:8080/swagger/index.html
```

## معماری

- **Handlers**: مدیریت درخواست‌های HTTP
- **Services**: منطق کسب‌وکار (OTP، rate limiting، JWT)
- **Repositories**: دسترسی به دیتابیس
- **Models**: ساختارهای داده
- **Middleware**: احراز هویت و محدودیت نرخ

## نکات امنیتی

- JWT secret باید قوی و منحصر به فرد باشد
- در production از HTTPS استفاده کنید
- Rate limiting برای جلوگیری از حملات brute force
- OTP ها فقط 2 دقیقه معتبر هستند
