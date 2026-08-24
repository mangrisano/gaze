package main

// splitArgs splits args at the first "--" into the watcher's own arguments and
// the command to run. With no "--", command is nil.
func splitArgs(args []string) (own []string, command []string) {
	for i, arg := range args {
		if arg == "--" {
			own = append(own, args[:i]...)
			command = append(command, args[i+1:]...)
			return own, command
		}
	}
	return args, nil
}
