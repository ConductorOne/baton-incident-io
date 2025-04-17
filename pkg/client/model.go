package client

type UserResponse struct {
	Users []User `json:"users"`
	Meta  Meta   `json:"pagination_meta"`
}

type ScheduleResponse struct {
	Schedule []Schedule `json:"schedules"`
	Meta     Meta       `json:"pagination_meta"`
}

type SingleUserResponse struct {
	User User `json:"user"`
}

type User struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Email       string     `json:"email"`
	BaseRole    BaseRole   `json:"base_role"`
	CustomRoles []BaseRole `json:"custom_roles,omitempty"`
}

type BaseRole struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Slug        string `json:"slug"`
}

type Meta struct {
	PageSize int    `json:"page_size"`
	After    string `json:"after"`
}

type Schedule struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	CurrentShifts []CurrentShift `json:"current_shifts"`
	Config        ScheduleConfig `json:"config"`
}

type ScheduleConfig struct {
	Rotation []Rotation `json:"rotations"`
}

type Rotation struct {
	ID    string      `json:"id"`
	Name  string      `json:"name"`
	Users []ShiftUser `json:"users"`
}

type ListScheduleResponse struct {
	Schedule []Schedule `json:"schedules"`
}

type CurrentShift struct {
	RotationID string    `json:"rotation_id"`
	User       ShiftUser `json:"user"`
	StartAt    string    `json:"start_at"`
	EndAt      string    `json:"end_at"`
}

type ShiftUser struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}
