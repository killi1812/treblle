package dto

type UserDto struct {
	Uuid      string `json:"uuid"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	OIB       string `json:"oib"`
	Residence string `json:"residence"`
	BirthDate string `json:"birthDate"`
	Email     string `json:"email"`
	Role      string `json:"role"`
}

// func (dto UserDto) ToModel() (*model.User, error) {
// 	uuid, err := uuid.Parse(dto.Uuid)
// 	if err != nil {
// 		zap.S().Error("Failed to parse uuid = %s, err = %+v", dto.Uuid, err)
// 		return nil, cerror.ErrBadUuid
// 	}
//
// 	bod, err := time.Parse(format.DateFormat, dto.BirthDate)
// 	if err != nil {
// 		zap.S().Errorf("Failed to parse BirthDate = %s, err = %+v", dto.BirthDate, err)
// 		return nil, cerror.ErrBadDateFormat
// 	}
//
// 	role, err := model.StrToUserRole(dto.Role)
// 	if err != nil {
// 		zap.S().Errorf("Failed to parse role = %+v, err = %+v", dto.Role, err)
// 		return nil, cerror.ErrUnknownRole
// 	}
//
// 	return &model.User{
// 		Uuid:      uuid,
// 		FirstName: dto.FirstName,
// 		LastName:  dto.LastName,
// 		OIB:       dto.OIB,
// 		Residence: dto.Residence,
// 		BirthDate: bod,
// 		Email:     dto.Email,
// 		Role:      role,
// 	}, nil
// }
//
// // FromModel returns a dto from model struct
// func (UserDto) FromModel(m *model.User) UserDto {
// 	dto := &UserDto{
// 		Uuid:      m.Uuid.String(),
// 		FirstName: m.FirstName,
// 		LastName:  m.LastName,
// 		OIB:       m.OIB,
// 		Residence: m.Residence,
// 		BirthDate: m.BirthDate.Format(format.DateFormat),
// 		Email:     m.Email,
// 		Role:      fmt.Sprint(m.Role),
// 	}
// 	return *dto
// }
