Force COPY
==========

```
fcopy SRC... DST
```

`fcopy` is the file copying program running on Command Prompt.
When `fcopy` fails to copy, it retries by methods below.

`The process cannot access the file because it is being used by another process.`
------------

If `DST` is used by another process, `fcopy` tries to rename busy `DST` to `DST-YYYYMMDD_hhmmss` and retry copying.


`Access is denied.`
---------------

If `DST` is the directory which not-administrator can not write (for example, `C:\Program Files`) , `fcopy` tries to show User-Account-Control dialog and run itself as Administrator.
