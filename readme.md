# Read out memory information for a process

The provided application supports the investigation of the memory behavior of Android processes.

## Run the application

Before running the application make sure all modules are available and updated
```
go mod tidy
```
To debug the application just use the `go run` statement
```
go run main.go <your arguments>
```
To compile and run the application use the ´go build´ statement
```
go build main.go
```

## Get detailed memory info

The main function of the application is to provide ( and if desired, store) memory information for Android applications. Below we describe the call to perform a test. The package name of the application is required for execution. This can be determined using `packages` (see below). It is displayed whether the application is currently running on the end device or is being terminated. In the event that the application has been terminated, the output and recording of memory information also ends. Both are automatically resumed when the Android app has been restarted.

```
go run main.go meminfo
```

The following options are provided:

```
- p string [mandatory]
    Package name of application to be analyzed

-adb string [optional]
    Absolute path of adb application (default: adb)

- t integer [optional]
    The refresh rate in seconds (default: 0)
    If no refresh rate is provided a single measurement is done

- f string [optional]
    A display filter for the results to be shown
    If no filter is provided all possible key value pairs are shown
    The filter value must be provided as regular expression

- o string [optional]
    The filename of a csv file the results are written to

- of string [optional]
    A filter for the results to be stored in the csv file
    If no filter is provided all possible key value pairs are stored
    The filter value must be provided as regular expression
```
Example: Return all memory information by using the provided adb application and filter the result in order to show and store only keys containg `Rss`. All results are stored within the `test.csv` file. 

```
go run main.go meminfo -t=1 -p="org.qtproject.example" -adb="C:\Users\Test\AppData\Local\Android\Sdk\platform-tools\adb" -f="Rss" -o test.csv -of="Rss"

Start:               09.03.2023 (22:36:25)
Current measurement: 09.03.2023 (22:36:30) [6]
Next measurement:    09.03.2023 (22:36:31) [7; rate=1/1sec]
Process:             ACTIVE (org.qtproject.example => pid:23428)

Name                       |    Value |       T1000 |        T100 |         T10 |          T1
---------------------------+----------+-------------+-------------+-------------+-------------
MEM Unknown Rss Total      |      972 |         0.0 |         0.0 |         0.0 |         0.0
MEM TOTAL Rss Total        |    71864 |         0.0 |         0.2 |         1.5 |         4.0
MEM Stack Rss Total        |      544 |         0.0 |         0.0 |         0.0 |         0.0
MEM Other mmap Rss Total   |      868 |         0.0 |         0.0 |         0.0 |         0.0
MEM Other dev Rss Total    |      228 |         0.0 |         0.0 |         0.0 |         0.0
MEM Native Heap Rss Total  |    22944 |         0.0 |         0.0 |         0.0 |         0.0
MEM Dalvik Other Rss Total |     1820 |         0.0 |         0.2 |         1.4 |         8.0
MEM Dalvik Heap Rss Total  |     3852 |         0.0 |         0.0 |         0.0 |         0.0
MEM Ashmem Rss Total       |        8 |         0.0 |         0.0 |         0.0 |         0.0
MEM .ttf mmap Rss Total    |      136 |         0.0 |         0.0 |         0.0 |         0.0
MEM .so mmap Rss Total     |    55332 |         0.0 |         0.0 |         0.0 |         0.0
MEM .oat mmap Rss Total    |     1988 |         0.0 |         0.0 |         0.0 |         0.0
MEM .jar mmap Rss Total    |    27024 |         0.1 |         1.2 |         9.3 |         0.0
MEM .dex mmap Rss Total    |      100 |         0.0 |         0.0 |         0.0 |         0.0
MEM .art mmap Rss Total    |    13972 |         0.0 |         0.0 |         0.0 |         0.0
MEM .apk mmap Rss Total    |    40852 |         0.0 |         0.0 |         0.0 |         0.0
APP Unknown: Rss(KB)       |     3636 |         0.0 |         0.0 |         0.0 |         0.0
APP Stack: Rss(KB)         |      544 |         0.0 |         0.0 |         0.0 |         0.0
APP Native Heap: Rss(KB)   |    22944 |         0.0 |         0.0 |         0.0 |         0.0
APP Java Heap: Rss(KB)     |    17824 |         0.0 |         0.0 |         0.0 |         0.0
APP Graphics: Rss(KB)      |        0 |         0.0 |         0.0 |         0.0 |         0.0
APP Code: Rss(KB)          |   125692 |         0.1 |         1.4 |        10.7 |         8.0

Addded 127 symbols to test.csv (filesize: 762B)
```
The content of the the `test.csv` is the following:
```
Time;MEM Unknown Rss Total;MEM TOTAL Rss Total;MEM Stack Rss Total;MEM Other mmap Rss Total;MEM Other dev Rss Total;MEM Native Heap Rss Total;MEM Dalvik Other Rss Total;MEM Dalvik Heap Rss Total;MEM Ashmem Rss Total;MEM .ttf mmap Rss Total;MEM .so mmap Rss Total;MEM .oat mmap Rss Total;MEM .jar mmap Rss Total;MEM .dex mmap Rss Total;MEM .art mmap Rss Total;MEM .apk mmap Rss Total;APP Unknown: Rss(KB);APP Stack: Rss(KB);APP Native Heap: Rss(KB);APP Java Heap: Rss(KB);APP Graphics: Rss(KB);APP Code: Rss(KB)
2023-03-09 22:37:58;13972;17824;1988;27024;100;0;22972;136;3852;22972;8;3632;71900;544;55332;125704;40852;544;1828;972;228;868
2023-03-09 22:37:59;13972;17824;1988;27024;100;0;22972;136;3852;22972;8;3636;71904;544;55332;125704;40852;544;1832;972;228;868
2023-03-09 22:38:00;13972;17824;1988;27024;100;0;22972;136;3852;22972;8;3636;71921;544;55332;125732;40852;544;1860;972;228;868
2023-03-09 22:38:01;13972;17824;1988;27024;100;0;22972;136;3852;22972;8;3636;71921;544;55332;125736;40852;544;1864;972;228;868
2023-03-09 22:38:02;13972;17824;1988;27024;100;0;22972;136;3852;22972;8;3636;71920;544;55332;125736;40852;544;1864;972;228;868
2023-03-09 22:38:03;13972;17824;1988;27024;100;0;22972;136;3852;22972;8;3636;71920;544;55332;125736;40852;544;1864;972;228;868
2023-03-09 22:38:04;13972;17824;1988;27024;100;0;22972;136;3852;22972;8;3632;71916;544;55332;125736;40852;544;1860;972;228;868
2023-03-09 22:38:05;13972;17824;1988;27024;100;0;22972;136;3852;22972;8;3636;71920;544;55332;125736;40852;544;1864;972;228;868
...
```
Example: Return all memory information by using the provided adb application and filter the result in order to show and store only keys containg `Rss`. All results are stored within the `test.csv` file. In this case the Android App is currently **not running**. In case the app is started while this tool runs it starts outputting the required information automatically.

```
go run main.go meminfo -t=1 -p="org.qtproject.example" -adb="C:\Users\Test\AppData\Local\Android\Sdk\platform-tools\adb" -f="Rss" -o test.csv -of="Rss"

Start:               09.03.2023 (22:47:48)
Current measurement: 09.03.2023 (22:49:15) [88]
Next measurement:    09.03.2023 (22:49:16) [89; rate=1/1sec]
Process:             INACTIVE since 86sec (org.qtproject.example)
```

### Rough trends

Rough trends of memory development are given for the application. These are not real trend lines, but the change of differently damped low-pass filters.

1. **T1000**: The change of the low-pass filter affected by the new value as follows: `filtervalue = filtervalue * 0.999 + newvalue * 0.001` (0.1%)
2. **T100**: The change of the low-pass filter affected by the new value as follows: `filtervalue = filtervalue * 0.99 + newvalue * 0.01` (1%)
3. **T10**: The change of the low-pass filter affected by the new value as follows: `filtervalue = filtervalue * 0.9 + newvalue * 0.1` (10%)
4. **T1**: The change of the value, compared to the last measured one

These trends are only displayed visually and are not saved. The processing of csv files, for example via the Excel application, offers far more possibilities for analysis.

## Get name of running processes

For the analysis of the memory behavior, the package name of the application to be examined must be known. This is specified according to the scheme `com.yourcompany.yourapp`. If you do not know the package name, you can list all currently running packages by using the `packages` argument.

```
go run main.go packages
```

The following options are provided:

```
-adb string [optional]
    Absolute path of adb application (default: adb)
```

Example: Output all processes by using the provided adb application

```
go run main.go packages -adb="C:\Users\Test\AppData\Local\Android\Sdk\platform-tools\adb"

Measurement:  09.03.2023 (22:02:42)
Packages:
   - zygote
   - wpa_supplicant
   - wificond
   - wifi_forwarder
   - webview_zygote
   - vold
   - ueventd
   - traced_probes
...
```

Example: Return all processes by using the default adb

```
go run main.go packages

Measurement:  09.03.2023 (22:08:03 )
Packages:
   - zygote
   - wpa_supplicant
   - wificond
   - wifi_forwarder
   - webview_zygote
   - vold
   - ueventd
   - traced_probes
...
```
## Get keys of memory information

In case a filter is to be applied to the screen or file output when `meminfo` is called, the keys provided by `adb` for the package can be output.
```
go run main.go names
```

The following options are provided:

```
- p string [mandatory]
    Package name of application to be analyzed

-adb string [optional]
    Absolute path of adb application (default: adb)
```

Example: Output all keys that are logable by `meminfo` for the package `qor.qtproject.example` by using the adb application provided.

```
adb ram reader> go run main.go names -p="org.qtproject.example" -adb="C:\Users\Test\AppData\Local\Android\Sdk\platform-tools\adb"                                      

Measurement:  09.03.2023 (22:52:33)
Process:      INACTIVE (org.qtproject.example)
```

Example: Output all keys that are logable by `meminfo` for the package `qor.qtproject.example` by using the adb application provided by `PATH`.

```
C:\GoProjects\src\adb ram reader> go run main.go names -p="org.qtproject.example"

Measurement:  09.03.2023 (23:12:17)
Process:      ACTIVE (org.qtproject.example => pid:6889)
Keywords:
   - SQL PAGECACHE_OVERFLOW
   - SQL MEMORY_USED
   - SQL MALLOC_SIZE
   - OBJECTS WebViews
   - OBJECTS Views
   - OBJECTS ViewRootImpl
   - OBJECTS Proxy Binders
   - OBJECTS Parcel memory
...
```

Example: Output all keys that are logable by `meminfo` for the package `qor.qtproject.example` by using the adb application provided. In this case the Android App is currently **not running**.

```
adb ram reader> go run main.go names -p="org.qtproject.example" -adb="C:\Users\Test\AppData\Local\Android\Sdk\platform-tools\adb"                                      

Measurement:  09.03.2023 (22:52:33)
Process:      INACTIVE (org.qtproject.example)
```
## Contact us
[SSE - Secure Systems Engineering GmbH](https://www.securesystems.de) supports companies in the production, maintenance and testing of secure systems and infrastructures.