package system_availability

var availabilityController = &AvailabilityController{}

func GetAvailabilityController() *AvailabilityController {
	return availabilityController
}
