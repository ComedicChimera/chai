# Standard IO

```
func print<...S: Show>(...s: S)
func println<..S: Show>(...s: S)
func printf<...T>(fmt: string, ...args: T)

func error<...S: Show>(...s: S)
func errorln<...S: Show>(...s: S)
func errorf<...T>(fmt: string, ...args: T)

func scan() Result[rune, Error]
func scanln() Result[string, Error]
func scanf<...T>(fmt: string, ...args: &T) Result[(), Error]

func flush() bool
