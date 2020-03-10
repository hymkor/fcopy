fcopy
=====

`fcopy` is the file copying program running on Command Prompt.
When `fcopy` fails to copy, it retries by methods below.

```
fcopy SRC... DST
```

-----

```
The process cannot access the file because it is being used by another process.
```

If `DST` is used by another process, rename `DST` to `DST-YYYYMMDD_hhmmss` and retry.


```
Access is denied.
```

If `DST` is the directory which not administrator can not write, switch user to administrator and retry.
