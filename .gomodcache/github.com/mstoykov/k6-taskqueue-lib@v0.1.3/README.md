## Introduction
[![Go Reference](https://pkg.go.dev/badge/github.com/mstoykov/k6-taskqueue-lib.svg)](https://pkg.go.dev/github.com/mstoykov/k6-taskqueue-lib)

This is a small task queuer for k6's event loop. 

This is useful when you will need to queue multiple tasks to the event loop asynchronously from one another. 

While the event loop has a way to queue a task it requires that you are currently executing *in* the event loop. If you need to queue another callback after that you need to re-register a new one while on the event loop. 

This isn't that hard but gets kind of complicated if you have to queue one each time an *asynchronous* message from a websocket comes for example. In those cases you need to have the previous callback re-register/refresh callback before/after doing w/e it needed on the event loop, as usual. But also have code to wait for this "refresh" and then execute the code you wanted - the one actually about the message you received, again re-registering for the next callback. This is more or less what the TaskQueue does. 

This is also useful even if you just want to implement setInterval and want to support queueing multiple executions if one starts to slow down. 

It's important to note that calling the Close method *is required* in order for the event loop to finish and end a k6 iteration.
