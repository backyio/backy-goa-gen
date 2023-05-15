# Logrus Plugin

The `Logrus` plugin is a [Goa](https://github.com/goadesign/goa/tree/v3) plugin
that adapt the basic logger to use the [logrus](https://github.com/sirupsen/logrus) library.

## Enabling the Plugin

To enable the plugin import it in your design.go file using the blank identifier `_` as follows:

```go

package design

import . "goa.design/goa/v3/http/design"
import . "goa.design/goa/v3/http/dsl"
import _ "github.com/backyio/backy-goa-gen/logrus" // Enables the plugin

var _ = API("...

```

and generate as usual:

```bash
goa gen PACKAGE
goa example PACKAGE
```

where `PACKAGE` is the Go import path of the design package.
