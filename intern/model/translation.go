package model

type Translation struct {
	Greeting       string `json:"greeting"`
	WelcomeMessage string `json:"welcome_message"`
	GuestForm TranslationGuestForm `json:"guest_form"`
}

type TranslationGuestForm struct {
	LabelInputFirstname string `json:"label_input_firstname"`
	LabelInputLastname string `json:"label_input_lastname"`
	LabelSelectDiet string  `json:"label_select_diet"`
	LabelButtonSubmit string  `json:"label_button_submit"`
	SelectOptionsDiet []string `json:"select_options_diet"`
}
