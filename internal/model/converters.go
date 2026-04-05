package model

func (u User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		Name:      u.Name,
		Email:     u.Email,
		Phone:     u.Phone,
		Role:      u.Role,
	}
}

func (s Service) ToResponse() ServiceResponse {
	return ServiceResponse{
		ID:              s.ID,
		CreatedAt:       s.CreatedAt,
		UpdatedAt:       s.UpdatedAt,
		Name:            s.Name,
		DurationMinutes: s.DurationMinutes,
		Price:           s.Price,
	}
}

func (b Booking) ToResponse() BookingResponse {
	return BookingResponse{
		ID:          b.ID,
		CreatedAt:   b.CreatedAt,
		UpdatedAt:   b.UpdatedAt,
		CustomerID:  b.CustomerID,
		EmployeeID:  b.EmployeeID,
		ServiceID:   b.ServiceID,
		BookingDate: b.BookingDate,
		Status:      b.Status,
		Notes:       b.Notes,
		Customer:    b.Customer.ToResponse(),
		Employee:    b.Employee.ToResponse(),
		Service:     b.Service.ToResponse(),
	}
}

func (s Schedule) ToResponse() ScheduleResponse {
	return ScheduleResponse{
		ID:         s.ID,
		CreatedAt:  s.CreatedAt,
		UpdatedAt:  s.UpdatedAt,
		EmployeeID: s.EmployeeID,
		WeekDay:    s.WeekDay,
		StartTime:  s.StartTime,
		EndTime:    s.EndTime,
		Employee:   s.Employee.ToResponse(),
	}
}

func (so ScheduleOverride) ToResponse() ScheduleOverrideResponse {
	return ScheduleOverrideResponse{
		ID:          so.ID,
		CreatedAt:   so.CreatedAt,
		UpdatedAt:   so.UpdatedAt,
		EmployeeID:  so.EmployeeID,
		Date:        so.Date,
		Reason:      so.Reason,
		AllDay:      so.AllDay,
		StartTime:   so.StartTime,
		EndTime:     so.EndTime,
		IsAvailable: so.IsAvailable,
		Employee:    so.Employee.ToResponse(),
	}
}

func (ts TimeSlot) ToResponse() TimeSlotResponse {
	return TimeSlotResponse{
		ID:          ts.ID,
		CreatedAt:   ts.CreatedAt,
		UpdatedAt:   ts.UpdatedAt,
		EmployeeID:  ts.EmployeeID,
		StartTime:   ts.StartTime,
		EndTime:     ts.EndTime,
		IsAvailable: ts.IsAvailable,
		Employee:    ts.Employee.ToResponse(),
	}
}
