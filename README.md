# Vehicle E-commerce & Multi-Branch Management System
## ប្រព័ន្ធពាណិជ្ជកម្មអេឡិចត្រូនិកយានយន្ត និងគ្រប់គ្រងច្រើនសាខា

---

## 🚀 Overview & Vision (ទស្សនវិស័យទូទៅ)
ប្រព័ន្ធពាណិជ្ជកម្មអេឡិចត្រូនិកទំនើប និងប្រព័ន្ធគ្រប់គ្រងសាខា (Multi-Branch ERP & E-Commerce) សម្រាប់លក់រថយន្ត ម៉ូតូ កង់ (Gasoline & Electric Vehicles) និងគ្រឿងបន្លាស់ ដែលគាំទ្រការលក់បង់ផ្ដាច់ (Full Payment), បង់រំលោះ (Installment Loan), ប្ដូរយានយន្តចាស់យកថ្មី (Trade-in), និងការទិញចូល-លក់ចេញ (Buy & Resell)។

---

## 🛠 Technology Stack

| Layer | Technologies |
| :--- | :--- |
| **Frontend** | Next.js 15, TypeScript, Tailwind CSS, Lucide Icons, Glassmorphism UI |
| **Backend** | Go (Golang 1.27.1+), Gin Web Framework, JWT RBAC, pq (PostgreSQL Driver) |
| **Database** | **Supabase (Cloud PostgreSQL)** - No Docker required! |
| **Notifications** | Telegram Bot API (Instant Alerts for Orders, Trade-In, Loans & Low Stock) |

---

## ☁️ ការរៀបចំ Supabase Database (Supabase Setup Guide)

គម្រោងនេះប្រើប្រាស់ **Supabase** ជា Cloud Database ដោយមិនចាំបាច់ប្រើប្រាស់ Docker ទៀតឡើយ៖

### ជំហានទី ១៖ បង្កើត Project និងរៀបចំ Schema លើ Supabase
1. ចូលទៅកាន់ [https://supabase.com](https://supabase.com) រួចបង្កើតគណនី ឬ **New Project**។
2. ក្នុង Project Dashboard ជ្រើសរើស menu **SQL Editor** នៅខាងឆ្វេង -> ចុច **New query**។
3. បើក file `backend/migrations/supabase_setup.sql` រួច Copy កូដ SQL ទាំងអស់ យកទៅ Paste ក្នុង SQL Editor ហើយចុច **Run**។
   *(វានឹងបង្កើត Table ទាំងអស់ និងបញ្ចូលទិន្នន័យគំរូរួមមាន រថយន្ត Toyota 2026-2027 ចំនួន ២៧ ម៉ូឌែល, សាខា, និង User គ្រប់គ្រងដោយស្វ័យប្រវត្តិ)*
   បន្ទាប់ពី setup នេះដំណើរការជោគជ័យ Backend នឹងរំលង schema និង seed migrations មូលដ្ឋាន ហើយអនុវត្តតែ migrations បន្ថែមសម្រាប់ Admin ERP។

### ជំហានទី ២៖ កំណត់ Connection String ក្នុង `.env`
1. ក្នុង Supabase Dashboard ចូលទៅកាន់ **Project Settings** > **Database** > ផ្នែក **Connection string** > ជ្រើសយក tab **URI**។
2. Copy URI នោះហើយយកទៅដាក់ក្នុង file `backend/.env`៖
```env
DATABASE_URL=postgresql://postgres:[YOUR-PASSWORD]@db.[YOUR-PROJECT-REF].supabase.co:5432/postgres?sslmode=require
```
*(សូមប្តូរ `[YOUR-PASSWORD]` ទៅជា password ដែលអ្នកបានកំណត់ពេលបង្កើត project នៅ supabase)*

កំណត់ `JWT_SECRET` ក្នុង `backend/.env` ជាខ្សែអក្សរសម្ងាត់ចៃដន្យយ៉ាងតិច ៣២ តួអក្សរ។ Backend នឹងមិនចាប់ផ្ដើមទេ ប្រសិនបើអថេរនេះមិនមាន។

---

## ⚡ One-Command Startup (បញ្ជាតែមួយដំណើរការប្រព័ន្ធទាំងមូល)

ដោយសារតែ Database ស្ថិតនៅលើ Cloud (Supabase) លោកអ្នកពុំបាច់បើក Docker ឡើយ! គ្រាន់តែ run command មួយគត់៖
> ត្រូវដំឡើង Go និង Node.js មុន។ ពេលចាប់ផ្ដើមលើកដំបូង `start.bat` នឹងដំឡើង Frontend dependencies ដោយស្វ័យប្រវត្តិ។

### វិធីទី ១៖ ប្រើ Batch File (ងាយស្រួលបំផុត - Double-Click ឬ Terminal)
```cmd
.\start.bat
```
> ឬគ្រាន់តែ **Double-Click លើ file start.bat** នៅក្នុង Folder D:\Project\vehicle-commerce។

### វិធីទី ២៖ ប្រើ PowerShell
```powershell
.\start.ps1
```

### វិធីទី ៣៖ ប្រើ npm
```bash
npm start
```

### 🛑 របៀបបិទ Services ទាំងអស់វិញ (One-Command Stop)
```cmd
.\stop.bat
# ឬ
.\stop.ps1
```

---

## 🏁 Manual Start (ប្រសិនបើចង់ run ម្ដងមួយៗ)

### 1. Run Backend Locally (Port 8080)
```powershell
cd D:\Project\vehicle-commerce\backend
go run ./cmd/api
```
Backend API ដំណើរការនៅ: `http://localhost:8080/api/v1`

### 2. Run Frontend (Port 3000)
```powershell
cd C:\xampp\htdocs\vehicle-commerce\frontend
npm install # ត្រូវធ្វើម្ដងដំបូង ឬក្រោយពេល dependencies ផ្លាស់ប្ដូរ
npm run dev
```
បើក Browser នៅ: `http://localhost:3000`

- User storefront: `http://localhost:3000/`
- Admin dashboard: `http://localhost:3000/admin`

ចូល Admin ដោយប្រើ **username ឬ email** របស់គណនីដែលមានតួនាទី `SUPER_ADMIN`, `BRANCH_ADMIN`, ឬ `STAFF`។ សម្រាប់ទិន្នន័យសាកល្បងដែលបានបញ្ចូលពី migrations ប្រើ `superadmin` (ឬ `admin@vehiclemall.com`) ជាមួយ password `admin123`។ គណនីទាំងនេះសម្រាប់សាកល្បងប៉ុណ្ណោះ—សូមប្តូរពាក្យសម្ងាត់ និងបិទគណនី demo មុនប្រើប្រាស់ពិត។

បើឃើញ `invalid credentials` សូមពិនិត្យថា username/email មាននៅក្នុងតារាង `users` ហើយពាក្យសម្ងាត់ត្រូវគ្នា។ អ៊ីមែលដែលមិនមានក្នុងតារាង `users` មិនអាចប្រើចូលបានទេ លុះត្រាតែបានបង្កើតគណនីនោះក្នុងប្រព័ន្ធ។

សំណើពី Checkout បង្កើត order ស្ថានភាព `PENDING` ដើម្បីឲ្យក្រុមការងារទាក់ទងបញ្ជាក់ព័ត៌មាន។ អ្នកទិញអាចជ្រើសទទួលនៅសាខា ឬដឹកដោយហាង។ សម្រាប់ការដឹកដោយហាង តម្លៃដឹកតាមទីតាំងមិនគិតជាមុនទេ៖ Admin ត្រូវបញ្ចូល និងបញ្ជាក់ថ្លៃក្នុង Order management ដើម្បីកែចំនួនសរុប មុនពេលទទួលការទូទាត់។ Wishlist រក្សាទុកក្នុង browser របស់អ្នកប្រើ។ ការកក់នេះមិនកាត់ស្តុករថយន្តភ្លាមៗទេ ព្រោះស្តុកពិត និងតម្លៃត្រូវបញ្ជាក់ដោយសាខាជាមុន។

---

## 🔑 Default Accounts (គណនីសាកល្បង)

| តួនាទី (Role) | Username / Email | Password | សិទ្ធិ និងការងារ |
| :--- | :--- | :--- | :--- |
| **Super Admin** | `superadmin` / `admin@vehiclemall.com` | `admin123` | គ្រប់គ្រងសាខាទាំងអស់, របាយការណ៍រួម |
| **Branch Manager** | `pp_manager` / `branch.pp@vehiclemall.com` | `admin123` | គ្រប់គ្រងសាខាភ្នំពេញ, ស្តុក, បុគ្គលិក |
| **Staff Sales** | `staff_dara` / `dara.sales@vehiclemall.com` | `admin123` | បង្កើត Order, ពិនិត្យ Trade-in & Loan |
| **Customer** | `customer_visal` / `visal.customer@gmail.com` | `admin123` | មើលទំនិញ, កក់រថយន្ត, ស្នើបង់រំលោះ |

---

## 🤖 Telegram Bot Configuration
1. បង្កើត Bot តាមរយៈ `@BotFather` នៅលើ Telegram ដើម្បីទទួលបាន `TELEGRAM_BOT_TOKEN`។
2. បង្កើត Telegram Group សម្រាប់បុគ្គលិក និងបន្ថែម Bot ចូល រួចយក `Chat ID` មកដាក់ក្នុង `.env`:
```env
TELEGRAM_BOT_TOKEN=123456789:ABCdefGhIJKlmNoPQRsTUVwxyZ
TELEGRAM_ALERT_CHAT_ID=-1001234567890
```
3. រាល់ពេលមានការកុម្ម៉ង់ថ្មី (New Order), សំណើ Trade-in, ឬស្តុកជិតអស់ ប្រព័ន្ធនឹងផ្ញើសារជូនដំណឹងភ្លាមៗ!
