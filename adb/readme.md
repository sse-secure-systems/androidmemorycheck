# The adp parsing part

To access the functions provided, a reader must first be created. When creating the reader, it is possible to specify where the `adb` application is located. 

```
reader := CreateReader("C:/Users/Test/AppData/Local/Android/Sdk/platform-tools/adb")
```
If `adb` is accessible via the `PATH` variable using `adb`, the path specification `adbpath` can be left empty:
```
reader := CreateReader("")
```
The 'Reader' interface now provides the following functions:
1. Read and parse memory info by using `adb shell dumpsys meminfo`
2. Read and parse running processes by using `adb shell ps`
3. Get rough trend information

## Get memory information

This is the core functionality of the application. It first determines the process ID of the provided `packagename` with the help of the `adb shell pidof` call. Using the process ID and the `adb shell dumpsys meminfo` command, the memory information is retrieved and parsed. For common information groups, the essential information is converted to a key-value pair.

```
Scan(packagename string) (map[string]int, string, error)
```

## Get running processes

The endpoint uses the `adb shell ps -A -o NAME` command to determine all active processes. It ignores process information enclosed by `[` and `]`.

```
Packages() []string
```

## Get rough trend

For the evaluation of the storage behavior, an initial indication of the trend of storage usage is required. For performance reasons, no complete history of all measured values is kept. Instead, a low-pass filter is used to determine trends.

```
Trend(key string, value int) (float64, float64, float64, float64)
```

The method is given a `key` and a current `value`. For the provided key it calculates four trends:
1. **result #1**: The change of the low-pass filter affected by the new `value` as follows: `filtervalue = filtervalue * 0.999 + value * 0.001` (0.1%)
2. **result #2**: The change of the low-pass filter affected by the new `value` as follows: `filtervalue = filtervalue * 0.99 + value * 0.01` (1%)
3. **result #3**: The change of the low-pass filter affected by the new `value` as follows: `filtervalue = filtervalue * 0.9 + value * 0.1` (10%)
4. **result #4**: The change of the `value`, compared to the last measured one

This is not an exact trend calculation, especially since the trend information is lower the "slower" a low-pass filter reacts. Nevertheless, rough trends allow an overview of the memory behavior.