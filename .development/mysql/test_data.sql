INSERT INTO service_logs VALUES (
  'api',
  '[2021-09-11T06:59:35.377Z]  "GET /" "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36"',
  'info',
  NOW(),
  NOW()
), (
  'web',
  '[2021-09-12T08:07:33.973Z]  "GET /favicon.ico" "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36"',
  'info',
  NOW(),
  NOW()
), (
  'web',
  'ingress-controller | 140562392353608:error:02001002:system library:fopen:No such file or directory:crypto/bio/bss_file.c:69:fopen(\'/ca.pem\',\'r\')',
  'fatal',
  NOW(),
  NOW()
), (
  'cache',
  '20:C 19 Sep 2021 00:01:55.730 * RDB: 0 MB of memory used by copy-on-write',
  'debug',
  NOW(),
  NOW()
), (
  'cache',
  '1:M 19 Sep 2021 00:25:59.878 * Background saving terminated with success',
  'debug',
  NOW(),
  NOW()
), (
  'cache',
  '1:M 19 Sep 2021 00:46:04.881 * DB saved on disk',
  'debug',
  NOW(),
  NOW()
);
