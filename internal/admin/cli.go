package admin

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const pageSize = 9

type AdminCLI struct {
	client *Client
	ui     *UI
	token  string
}

func NewCLI(serverURL string) *AdminCLI {
	return &AdminCLI{
		client: NewClient(serverURL),
		ui:     NewUI(),
	}
}

func (c *AdminCLI) Run() error {
	for {
		c.ui.ClearScreen()
		initialized, err := c.client.GetSetupStatus()
		if err != nil {
			return err
		}

		if initialized {
			break
		}

		fmt.Println("Admin setup is not complete.")
		fmt.Println()
		login := c.ui.Prompt("Login: ")
		password := c.ui.Prompt("Password: ")

		err = c.client.CreateAdmin(login, password)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to create admin:", err)
			c.ui.Pause()
			continue
		}

		fmt.Println("Admin added successfully.")
		c.ui.Pause()
	}

	for {
		c.ui.ClearScreen()
		fmt.Println("Admin login required.")
		fmt.Println()
		login := c.ui.Prompt("Login: ")
		password := c.ui.Prompt("Password: ")

		token, err := c.client.LoginAdmin(login, password)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Login failed:", err)
			c.ui.Pause()
			continue
		}

		c.token = token
		break
	}

	c.interactiveMenu()
	return nil
}

func (c *AdminCLI) interactiveMenu() {
	for {
		c.ui.ClearScreen()
		fmt.Println("Admin CLI")
		fmt.Println("=============")
		fmt.Println("1 - Список пользователей")
		fmt.Println("2 - Создать пользователя")
		fmt.Println("3 - Список сотрудников")
		fmt.Println("4 - Добавить сотрудника")
		fmt.Println("0 - Logout")
		choice := c.ui.Prompt("Введите номер: ")

		switch choice {
		case "1":
			c.listUsersMenu()
		case "2":
			c.createUserMenu()
		case "3":
			c.listEmployeesMenu()
		case "4":
			c.createEmployeeMenu()
		case "0":
			c.token = ""
			c.ui.ClearScreen()
			return
		default:
			fmt.Println("Неизвестная команда, попробуйте снова.")
			c.ui.Pause()
		}
	}
}

func (c *AdminCLI) listUsersMenu() {
	selected, done, err := browsePaginated(c, "Пользователи", 1,
		func(page int) ([]UserListItem, int, error) {
			resp, err := c.client.ListUsers(page, pageSize, c.token)
			if err != nil {
				return nil, 0, err
			}
			return resp.Users, resp.Total, nil
		},
		func(idx int, user UserListItem) {
			current := ""
			if user.IsCurrent {
				current = " (Вы)"
			}
			fmt.Printf("%d. %d %s (%s)%s\n", idx+1, user.UserID, user.Login, user.Role, current)
		},
		"0 - Назад | n - следующая страница | p - предыдущая страница | 1-9 - выбрать пользователя",
		func(choice string, page, totalPages int, users []UserListItem) (int, *UserListItem, bool, error) {
			switch choice {
			case "0":
				return page, nil, true, nil
			case "n":
				if page < totalPages {
					page++
				}
				return page, nil, false, nil
			case "p":
				if page > 1 {
					page--
				}
				return page, nil, false, nil
			default:
				idx, err := strconv.Atoi(choice)
				if err != nil || idx < 1 || idx > len(users) {
					fmt.Println("Неверный выбор, попробуйте снова.")
					c.ui.Pause()
					return page, nil, false, nil
				}

				selected := users[idx-1]
				if selected.IsCurrent {
					fmt.Println("Нельзя изменить или удалить самого себя.")
					c.ui.Pause()
					return page, nil, false, nil
				}

				return page, &selected, false, nil
			}
		})

	if err != nil {
		fmt.Fprintln(os.Stderr, "Ошибка загрузки списка пользователей:", err)
		c.ui.Pause()
		return
	}
	if done {
		return
	}

	c.manageUser(*selected)
}

func (c *AdminCLI) manageUser(user UserListItem) {
	for {
		c.ui.ClearScreen()
		fmt.Printf("Пользователь %d %s (%s)\n", user.UserID, user.Login, user.Role)
		fmt.Println("1 - Удалить пользователя")
		fmt.Println("2 - Сменить роль")
		fmt.Println("3 - Добавить посещение")
		fmt.Println("4 - Посмотреть посещения")
		fmt.Println("0 - Назад")
		choice := c.ui.Prompt("Выберите действие: ")

		switch choice {
		case "0":
			return
		case "1":
			confirm := c.ui.Prompt("Подтвердите удаление пользователя (y/n): ")
			if strings.ToLower(confirm) != "yes" && strings.ToLower(confirm) != "y" {
				fmt.Println("Удаление отменено.")
				c.ui.Pause()
				continue
			}

			if err := c.client.DeleteUser(user.UserID, c.token); err != nil {
				fmt.Fprintln(os.Stderr, "Ошибка удаления:", err)
			} else {
				fmt.Println("Пользователь удалён.")
			}
			c.ui.Pause()
			return
		case "2":
			newRole := c.ui.Prompt("Новая роль (admin/manager/employee): ")
			if err := c.client.UpdateUserRole(user.UserID, newRole, c.token); err != nil {
				fmt.Fprintln(os.Stderr, "Ошибка смены роли:", err)
			} else {
				fmt.Println("Роль обновлена.")
			}
			c.ui.Pause()
			return
		case "3":
			userDetails, err := c.client.GetUserByID(user.UserID, c.token)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Ошибка получения данных пользователя:", err)
				c.ui.Pause()
				continue
			}

			if userDetails.EmployeeID == nil {
				fmt.Println("Пользователь не связан с сотрудником.")
				c.ui.Pause()
				continue
			}

			if err := c.client.CreateAttendance(*userDetails.EmployeeID, c.token); err != nil {
				fmt.Fprintln(os.Stderr, "Ошибка добавления посещения:", err)
			} else {
				fmt.Println("Посещение добавлено.")
			}
			c.ui.Pause()
		case "4":
			c.viewUserAttendances(user)
		default:
			fmt.Println("Неизвестная команда, попробуйте снова.")
			c.ui.Pause()
		}
	}
}

func (c *AdminCLI) createUserMenu() {
	employeeID := c.selectEmployeeForNewUser()
	if employeeID <= 0 {
		return
	}

	c.ui.ClearScreen()
	fmt.Println("Создание нового пользователя")
	fmt.Println("===========================")
	login := c.ui.Prompt("Логин: ")
	password := c.ui.Prompt("Пароль: ")
	role := c.ui.Prompt("Роль (admin/manager/employee): ")

	if err := c.client.CreateUser(login, password, role, employeeID, c.token); err != nil {
		fmt.Fprintln(os.Stderr, "Ошибка создания пользователя:", err)
	} else {
		fmt.Println("Пользователь создан успешно.")
	}
	c.ui.Pause()
}

func (c *AdminCLI) createEmployeeMenu() {
	c.ui.ClearScreen()
	fmt.Println("Добавление сотрудника")
	fmt.Println("====================")
	firstName := c.ui.Prompt("Имя: ")
	lastName := c.ui.Prompt("Фамилия: ")
	position := c.ui.Prompt("Должность: ")

	if firstName == "" || lastName == "" {
		fmt.Fprintln(os.Stderr, "Имя и фамилия обязательны.")
		c.ui.Pause()
		return
	}

	if err := c.client.CreateEmployee(firstName, lastName, position, c.token); err != nil {
		fmt.Fprintln(os.Stderr, "Ошибка добавления сотрудника:", err)
	} else {
		fmt.Println("Сотрудник добавлен успешно.")
	}
	c.ui.Pause()
}

func (c *AdminCLI) listEmployeesMenu() {
	_, _, err := browsePaginated(c, "Сотрудники", 1,
		func(page int) ([]EmployeeListItem, int, error) {
			resp, err := c.client.ListEmployees(page, pageSize, c.token)
			if err != nil {
				return nil, 0, err
			}
			return resp.Employees, resp.Total, nil
		},
		func(idx int, employee EmployeeListItem) {
			fmt.Printf("%d. %d %s %s (%s)\n", idx+1, employee.EmployeeID, employee.LastName, employee.FirstName, employee.Position)
		},
		"0 - Назад | n - следующая страница | p - предыдущая страница",
		func(choice string, page, totalPages int, _ []EmployeeListItem) (int, *EmployeeListItem, bool, error) {
			switch choice {
			case "0":
				return page, nil, true, nil
			case "n":
				if page < totalPages {
					page++
				}
				return page, nil, false, nil
			case "p":
				if page > 1 {
					page--
				}
				return page, nil, false, nil
			default:
				fmt.Println("Неверный выбор, попробуйте снова.")
				c.ui.Pause()
				return page, nil, false, nil
			}
		})

	if err != nil {
		fmt.Fprintln(os.Stderr, "Ошибка загрузки списка сотрудников:", err)
		c.ui.Pause()
	}
}

func (c *AdminCLI) viewUserAttendances(user UserListItem) {
	// Get user details to get employee_id
	userDetails, err := c.client.GetUserByID(user.UserID, c.token)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Ошибка получения данных пользователя:", err)
		c.ui.Pause()
		return
	}

	if userDetails.EmployeeID == nil {
		fmt.Println("Пользователь не связан с сотрудником.")
		c.ui.Pause()
		return
	}

	resp, err := c.client.GetAttendancesByEmployee(*userDetails.EmployeeID, c.token)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Ошибка загрузки посещений:", err)
		c.ui.Pause()
		return
	}

	if len(resp.Attendances) == 0 {
		fmt.Println("Нет посещений.")
		c.ui.Pause()
		return
	}

	for {
		c.ui.ClearScreen()
		fmt.Printf("Посещения пользователя %s\n", user.Login)
		fmt.Println("------------------------------------------------------")
		for idx, att := range resp.Attendances {
			fmt.Printf("%d. %s %s (%s)\n", idx+1, att.Date, att.TimeIn, att.Status)
		}
		fmt.Println("------------------------------------------------------")
		fmt.Println("0 - Назад | 1-9 - выбрать посещение для удаления")
		choice := c.ui.Prompt("Выберите: ")

		if choice == "0" {
			return
		}

		idx, err := strconv.Atoi(choice)
		if err != nil || idx < 1 || idx > len(resp.Attendances) {
			fmt.Println("Неверный выбор, попробуйте снова.")
			c.ui.Pause()
			continue
		}

		selected := resp.Attendances[idx-1]
		confirm := c.ui.Prompt(fmt.Sprintf("Подтвердите удаление посещения %s %s (y/n): ", selected.Date, selected.TimeIn))
		if strings.ToLower(confirm) == "yes" || strings.ToLower(confirm) == "y" {
			if err := c.client.DeleteAttendance(selected.AttendanceID, c.token); err != nil {
				fmt.Fprintln(os.Stderr, "Ошибка удаления посещения:", err)
			} else {
				fmt.Println("Посещение удалено.")
				resp, err = c.client.GetAttendancesByEmployee(*userDetails.EmployeeID, c.token)
				if err != nil {
					fmt.Fprintln(os.Stderr, "Ошибка обновления списка:", err)
					c.ui.Pause()
					return
				}
			}
		} else {
			fmt.Println("Удаление отменено.")
		}
		c.ui.Pause()
	}
}
func (c *AdminCLI) selectEmployeeForNewUser() int {
	selected, done, err := browsePaginated(c, "Сотрудники", 1,
		func(page int) ([]EmployeeListItem, int, error) {
			resp, err := c.client.ListEmployees(page, pageSize, c.token)
			if err != nil {
				return nil, 0, err
			}
			return resp.Employees, resp.Total, nil
		},
		func(idx int, employee EmployeeListItem) {
			fmt.Printf("%d. %d %s %s (%s)\n", idx+1, employee.EmployeeID, employee.LastName, employee.FirstName, employee.Position)
		},
		"0 - Отмена | a - Добавить нового сотрудника | n - след. страница | p - пред. страница",
		func(choice string, page, totalPages int, employees []EmployeeListItem) (int, *EmployeeListItem, bool, error) {
			switch choice {
			case "0":
				return page, nil, true, nil
			case "a":
				c.createEmployeeMenu()
				return page, nil, false, nil
			case "n":
				if page < totalPages {
					page++
				}
				return page, nil, false, nil
			case "p":
				if page > 1 {
					page--
				}
				return page, nil, false, nil
			default:
				idx, err := strconv.Atoi(choice)
				if err != nil || idx < 1 || idx > len(employees) {
					fmt.Println("Неверный выбор, попробуйте снова.")
					c.ui.Pause()
					return page, nil, false, nil
				}
				selected := employees[idx-1]
				return page, &selected, false, nil
			}
		})

	if err != nil {
		fmt.Fprintln(os.Stderr, "Ошибка загрузки списка сотрудников:", err)
		c.ui.Pause()
		return 0
	}
	if done {
		return 0
	}

	return selected.EmployeeID
}

func browsePaginated[T any](c *AdminCLI, title string, page int, fetch func(int) ([]T, int, error), render func(int, T), footer string, handleChoice func(string, int, int, []T) (int, *T, bool, error)) (*T, bool, error) {
	for {
		c.ui.ClearScreen()
		items, total, err := fetch(page)
		if err != nil {
			return nil, false, err
		}

		totalPages := c.pageCount(total)
		fmt.Printf("%s — страница %d из %d, всего %d\n", title, page, totalPages, total)
		fmt.Println("------------------------------------------------------")
		for idx, item := range items {
			render(idx, item)
		}
		fmt.Println("------------------------------------------------------")
		fmt.Println(footer)
		choice := c.ui.Prompt("Выберите: ")

		nextPage, selected, done, err := handleChoice(choice, page, totalPages, items)
		if err != nil {
			return nil, false, err
		}
		if selected != nil {
			return selected, false, nil
		}
		if done {
			return nil, true, nil
		}
		page = nextPage
	}
}

func (c *AdminCLI) pageCount(total int) int {
	pages := (total + pageSize - 1) / pageSize
	if pages == 0 {
		pages = 1
	}
	return pages
}
