# Standard IO

```
func Print<...S: Show>(...s: S)
func Println<..S: Show>(...s: S)
func Printf<...T>(fmt: string, ...args: T)

func Error<...S: Show>(...s: S)
func Errorln<...S: Show>(...s: S)
func Errorf<...T>(fmt: string, ...args: T)

func Scan() Result<rune, Error>
func Scanln() Result<string, Error>
func Scanf<...T>(fmt: string, ...args: &T) Result<(), Error>

func Flush() bool
