Analyze the k6 test script and extract the extensions that the script depends on.

### Sources

Dependencies can come from three sources: k6 test script, manifest file, `K6_DEPENDENCIES` environment variable. Instead of these three sources, a k6 archive can also be specified, which can contain all three sources (currently two actually, because the manifest file is not yet included in the k6 archive). An archive is a tar file, which can be created using the k6 archive command.

> *NOTE*: It is assumed that the script and all dependencies are in the archive. No external dependencies are analyzed.

The name of k6 test script or archive can be specified as the positional argument in the command invocation. Alternatively, the content can be provided in the stdin. If stdin is used, the input format ('js' for script of or 'tar' for archive) must be specified using the `--input` parameter.

Primarily, the k6 test script is the source of dependencies. The test script and the local and remote JavaScript modules it uses are recursively analyzed. The extensions used by the test script are collected. In addition to the require function and import expression, the `"use k6 ..."` directive can be used to specify additional extension dependencies. If necessary, the `"use k6 ..."` directive can also be used to specify version constraints.

    "use k6 > 0.54";
    "use k6 with k6/x/faker > 0.4.0";
    "use k6 with k6/x/sql >= 1.0.1";

    import { Faker } from "k6/x/faker";
    import sql from "k6/x/sql";
    import driver from "k6/x/sql/driver/ramsql";

Dependencies and version constraints can also be specified in the so-called manifest file. The default name of the manifest file is `package.json` and it is automatically searched from the directory containing the test script up to the root directory. The `dependencies` property of the manifest file contains the dependencies in JSON format.

    {"dependencies":{"k6":">0.54","k6/x/faker":">0.4.0","k6/x/sql":>=v1.0.1"}}

Dependencies and version constraints can also be specified in the `K6_DEPENDENCIES` environment variable. The value of the variable is a list of dependencies in a one-line text format.

    k6>0.54;k6/x/faker>0.4.0;k6/x/sql>=v1.0.1

### Format

By default, dependencies are written as a JSON object. The property name is the name of the dependency and the property value is the version constraints of the dependency.

    {"k6":">0.54","k6/x/faker":">0.4.0","k6/x/sql":">=1.0.1","k6/x/sql/driver/ramsql":"*"}

Additional output formats:

 * `text` - One line text format. A semicolon-separated sequence of the text format of each dependency. The first element of the series is `k6` (if there is one), the following elements follow each other in lexically increasing order based on the name.

        k6>0.54;k6/x/faker>0.4.0;k6/x/sql>=1.0.1;k6/x/sql/driver/ramsql*

 * `js` - A consecutive, one-line JavaScript string directives. The first element of the series is `k6` (if there is one), the following elements follow each other in lexically increasing order based on the name.

        "use k6>0.54";
        "use k6 with k6/x/faker>0.4.0";
        "use k6 with k6/x/sql>=v1.0.1";

### Output

By default, dependencies are written to standard output. By using the `-o/--output` flag, the dependencies can be written to a file.
