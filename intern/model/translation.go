package model

type Translation struct {
	Greeting       string `json:"greeting"`
	WelcomeMessage string `json:"welcome_message"`
	GuestForm TranslationGuestForm `json:"guest_form"`
	Location	TranslationLocationSection `json:"location"`
	Navigation TranslationNavigation `json:"navigation"`
}

type TranslationGuestForm struct {
	LabelInputFirstname string `json:"label_input_firstname"`
	LabelInputLastname string `json:"label_input_lastname"`
	LabelSelectDiet string  `json:"label_select_diet"`
	LabelChildInput string `json:"label_child_input"`
	LabelButtonSubmit string  `json:"label_button_submit"`
	SelectOptionsDiet []string `json:"select_options_diet"`
}

type TranslationLocationSection struct {
	Title string `json:"title"`
	OpenExternally string `json:"openExternally"`
}

type TranslationNavigation struct {
	Guests string `json:"guests"`
	Map string `json:"map"`
}
