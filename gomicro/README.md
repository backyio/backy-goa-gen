# Logrus Plugin

The `Micro logger` plugin is a [Goa](https://github.com/goadesign/goa/tree/v3) plugin
that adapt the logger to use the [gomicro logger](https://go-micro.dev) library.

## Enabling the Plugin

To enable the plugin import it in your design.go file using the blank identifier `_` as follows:

```go

package design

import . "goa.design/goa/v3/http/design"
import . "goa.design/goa/v3/http/dsl"
import _ "github.com/backyio/backy-goa-gen/gomicro" // Enables the plugin

var _ = API("...

```

and generate as usual:

```bash
goa gen PACKAGE
goa example PACKAGE
```

where `PACKAGE` is the Go import path of the design package.
