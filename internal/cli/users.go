package cli

import (
	"fmt"

	"flexphish/internal/auth"
)

func HandleUserCommands(opts *CLIOptions, service auth.Service) bool {

	if opts.CreateUser {

		if opts.Email == "" || opts.Password == "" {
			fmt.Println("email and password are required")
			return true
		}

		role := auth.RoleUser
		if opts.Role == "admin" {
			role = auth.RoleAdmin
		}

		user, err := service.Register(opts.Email, opts.Password, role)
		if err != nil {
			fmt.Println("Error:", err)
			return true
		}

		fmt.Println("User created:", user.Email)
		return true
	}

	if opts.DeleteUser {

		if opts.Email == "" {
			fmt.Println("email is required")
			return true
		}

		err := service.DeleteByEmail(opts.Email)
		if err != nil {
			fmt.Println("Error:", err)
			return true
		}

		fmt.Println("User deleted:", opts.Email)
		return true
	}

	return false
}
