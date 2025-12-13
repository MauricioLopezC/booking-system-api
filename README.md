# reservation-system-api

# To do
## Refactor bookings service:
In order to avoid inconsistencies when creating or updating a booking:
1. First get the available slots using `FindAvailableSlots` function
2. Check if the `bookingDate` of new booking match with available slots

To prevent two or more bookings on the same date time, create an unique index between the following attributes of Booking model: `EmployeeID` and `BookingDate`.



## Schedule Exceptions Service
- [ ] Continue implementing routing...


