# Penilaian apresiasi trainee

Instalasi pada database yang sudah memiliki `users` dan `stores`:

```powershell
Get-Content -Raw -Encoding UTF8 migrations/000003_create_trainee_assessments.up.sql |
  & 'C:\xampp8.2.12\mysql\bin\mysql.exe' -u root --default-character-set=utf8mb4 hr_trainee_appreciation_assessment
Get-Content -Raw -Encoding UTF8 seeds/000006_trainee_assessment_rubric.sql |
  & 'C:\xampp8.2.12\mysql\bin\mysql.exe' -u root --default-character-set=utf8mb4 hr_trainee_appreciation_assessment
```

Jalankan migrasi sekali. Seed dapat dijalankan ulang tanpa menimpa rubrik yang sudah diedit.
SQL starter lama belum menyertakan modul ini; jalankan migrasi dan seed di atas setelah import starter.

## Tabel

| Tabel/view | Isi |
| --- | --- |
| trainee_assessment_aspects | Lima aspek form |
| trainee_assessment_criteria | Sepuluh subaspek, urutan dan bobot |
| trainee_assessment_score_rules | Batas jumlah temuan untuk tiap skor 1–5 |
| trainee_assessments | Karyawan, penilai, store, bagian, bulan pelatihan, tanggal, catatan dan tanda tangan |
| trainee_assessment_details | Nilai, jumlah temuan, bukti dan salinan rubrik tiap penilaian |
| trainee_assessment_results | Total nilai berbobot, bobot, rata-rata dan kategori hasil |

Karyawan dan penilai mengacu ke `users`, MK mengacu ke `stores`. Bagian disimpan sebagai teks karena belum ada master bagian. Nama, NIK, bagian, store dan nama penilai disimpan saat membuat form untuk menjaga histori.
`training_month` adalah bulan ke-berapa pelatihan, bukan bulan kalender. Diasumsikan satu penilaian per karyawan per bulan pelatihan; jika ada program pelatihan berulang, tambahkan entitas program dan sesuaikan unique key sebelum pemakaian.

## Menyimpan penilaian

Buat header dan salin seluruh kriteria ke detail dalam satu transaksi. Saat menyalin, isi `aspect_name`, `criterion_description`, dan `weight` dari master. Kolom skor boleh kosong pada draft. Gunakan jumlah temuan untuk memilih skor dari `trainee_assessment_score_rules`, atau simpan skor pilihan penilai.

Sebelum mengubah status ke `final`, aplikasi wajib memvalidasi bahwa tepat sepuluh kriteria rubrik tersedia, seluruh skor terisi, catatan terisi, dan penilai berhak mengakses store tersebut. Database memvalidasi rentang skor, bobot positif, bulan positif, relasi, dan catatan wajib pada status final; kelengkapan lintas baris dan otorisasi harus ditangani aplikasi.

Hasil view menggunakan `SUM(weight * score) / SUM(weight)`, dibulatkan dua desimal. Rata-rata belum tersedia jika detail kosong atau ada skor yang belum diisi. Kategori sesuai form: 1.00–1.80 Sangat Kurang Sekali, 1.81–2.60 Kurang Sekali, 2.61–3.40 Kurang, 3.41–4.20 Mendekati Standar, 4.21–5.00 Sesuai Standar.

Rubrik gambar kedua menggunakan istilah sedikit berbeda untuk FEFO dan sampling dibanding gambar pertama. Teks detail mengikuti gambar pertama, batas skor mengikuti gambar kedua. Simbol batas skor terbawah diinterpretasikan sebagai `>= 5` atau `>= 6` supaya seluruh jumlah temuan memiliki satu skor.

Kolom signature path adalah lokasi file privat; timestamp tanda tangan diisi setelah tindakan konfirmasi, bukan otomatis. Modul upload/tanda tangan dan halaman pengisian belum dibuat.

Rollback `000003_create_trainee_assessments.down.sql` menghapus seluruh tabel penilaian beserta datanya. Jangan jalankan pada database berisi penilaian yang perlu dipertahankan.
