package auth

func CreateNewUser() (bool, error)

func Authenticate() (bool, error)

func GetToken() (string, error)

func ValidateToken() (string, error)
