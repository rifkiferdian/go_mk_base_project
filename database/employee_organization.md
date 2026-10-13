# Data karyawan dan organisasi retail

Migrasi `000004_create_employee_organization.up.sql` memerlukan migrasi 000001–000003. Jalankan sekali, lalu jalankan seed `000010_employee_organization_dummy.sql` khusus development. Seed tidak menimpa biodata/password existing dan tidak memberikan permission admin kepada akun dummy.

Tabel: `job_levels`, `positions`, `departments`, `employees`, `employee_profiles`, `employee_assignments`, dan `employee_reporting_lines`.
`users.employee_id` nullable dan unik: satu karyawan maksimal mempunyai satu akun login; akun sistem tidak wajib terhubung karyawan. Akun admin existing tetap dipertahankan.

Penempatan aktif adalah baris dengan `ended_at IS NULL`. Satu karyawan memiliki satu penempatan aktif dan satu atasan langsung aktif. Saat mutasi/promosi, tutup penempatan dan reporting line lama lalu buat baris baru dalam transaksi. Aplikasi perlu memvalidasi periode tidak bertumpang tindih, masa berlaku hubungan berada dalam masa penempatan, atasan/bawahan bukan karyawan yang sama, dan tidak ada siklus hierarki. Foreign key serta check database hanya menjaga relasi dasar, bukan seluruh aturan hierarki.

Identitas karyawan bersumber dari `employees`; kolom nama/email/NIK lama di `users` masih dipertahankan karena modul User dan login existing memakainya. Integrasi/sinkronisasi halaman biodata belum dibuat.

Penilaian sekarang memiliki `trainee_employee_id` wajib, `trainee_assignment_id`, `assessor_assignment_id`, `trainee_position_name`, dan `assessor_position_name`. `trainee_user_id` lama tetap ada tetapi boleh kosong. Gunakan employee ID pada penilaian baru; `assessor_user_id` mencatat akun pengisi. Aplikasi perlu memastikan assignment memang milik trainee/penilai dan sesuai store/tanggal. Riwayat penilaian yang sudah ada dibackfill berdasarkan user lama; jika NIK kosong digunakan `LEGACY-USER-<id>` untuk diperiksa HR.

## Contoh fiktif MK5

| NIK | Nama | Jabatan | Atasan | Login |
| --- | --- | --- | --- | --- |
| DUMMY-MGR-001 | Budi Contoh | Store Manager | — | dummy.manager |
| DUMMY-SPV-001 | Sari Contoh | Supervisor Gudang | Budi Contoh | dummy.spv.gudang |
| DUMMY-SPV-002 | Dewi Contoh | Supervisor Kasir | Budi Contoh | dummy.spv.kasir |
| DUMMY-STF-001 | Andi Contoh | Staff Gudang | Sari Contoh | Tidak ada |
| DUMMY-TRN-001 | Rina Contoh | Trainee Gudang | Sari Contoh | Tidak ada |
| DUMMY-TRN-002 | Dimas Contoh | Trainee Kasir | Dewi Contoh | Tidak ada |

Password tiga akun dummy: `qweqwe`, hanya development. Role manager/supervisor belum mendapat permission penilaian karena modul transaksi penilaian belum dibuat. Akses penilaian nantinya harus memeriksa reporting line dan permission, bukan hanya tingkatan jabatan.

Tabel persetujuan manager belum ditambahkan: alur persetujuan masih opsional dan belum ditetapkan. Migrasi ini tidak menyediakan rollback otomatis karena menghapus master karyawan/relasi akun dapat merusak riwayat; gunakan backup atau migrasi koreksi yang mempertahankan data.
