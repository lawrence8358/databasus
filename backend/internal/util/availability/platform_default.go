//go:build !windows

package availability

func getPlatformUnsupportedEngines() []Engine {
	return nil
}

func getPlatformUnsupportedFeatures() []Feature {
	return nil
}
