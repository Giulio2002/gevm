package spec

// MaxCodeSizeForFork returns the EIP-170 runtime code limit, raised to
// 64 KiB by EIP-7954 in Amsterdam. Before Spurious Dragon the caller must
// not enforce a runtime code limit.
func MaxCodeSizeForFork(fork ForkID) int {
	if fork.IsEnabledIn(Amsterdam) {
		return 0x10000
	}
	return MaxCodeSize
}

// MaxInitcodeSizeForFork returns the EIP-3860 limit (enforced from Shanghai).
func MaxInitcodeSizeForFork(fork ForkID) int {
	return 2 * MaxCodeSizeForFork(fork)
}
