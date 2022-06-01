# Test cases

1. [test_00](./cases/test_00/README.md)
2. [test_01](./cases/test_01/README.md)
3. [test_02](./cases/test_02/README.md)
4. [test_03](./cases/test_03/README.md)
5. [test_04](./cases/test_04/README.md)
6. [test_05](./cases/test_05/README.md)

---

## test_00
---

Run all binaries without the config file.

---
## test_01
---

1. run gosler, save to gob
2. run gobler, render gob to file

---
## test_02
---

goslmailer runs with broken sacct line (-j jobid missing)

---
## test_03
---

goslmailer render msteams json to file (actual data)
Job start

---
## test_04
---

goslmailer render msteams json to file (actual data)
Job end - fail


## test_05
---

Test goslmailer on SLURM versions (<21.8.x) that don't set the job information in as env variables

---
