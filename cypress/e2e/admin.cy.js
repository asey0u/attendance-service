describe('Администратор', () => {

  beforeEach(() => {
    cy.loginFast('admin', 'admin')
    cy.visit('/admin')
  })

  it('дашборд показывает все счётчики', () => {
    cy.contains('Пользователи')
    cy.contains('Сотрудники')
    cy.contains('Отделы')
    cy.contains('Ожидают')
  })

  it('страница пользователей загружается с формой создания', () => {
    cy.visit('/admin/users')
    cy.contains('Пользователи')
    cy.get('form[action="/admin/users"] input[name=login]').should('exist')
    cy.get('form[action="/admin/users"] input[name=password]').should('exist')
    cy.get('form[action="/admin/users"] select[name=role]').should('exist')
  })

  it('поиск сотрудника в форме создания пользователя', () => {
    cy.visit('/admin/users')
    cy.intercept('GET', '/admin/users/employee-search*').as('empSearch')
    cy.get('#emp-search-input').type('Кузнецов')
    cy.wait('@empSearch')
    cy.get('#emp-results button').should('have.length.greaterThan', 0)
    cy.get('#emp-results button').first().click()
    cy.get('#emp-id-input').should('not.have.value', '')
    cy.get('#emp-selected').should('be.visible')
  })

  it('ошибка при создании пользователя с коротким логином', () => {
    cy.visit('/admin/users')
    cy.get('form[action="/admin/users"]').within(() => {
      cy.get('input[name=login]').type('ab')
      cy.get('input[name=password]').type('password123')
      cy.get('select[name=role]').select('employee')
    })
    cy.intercept('GET', '/admin/users/employee-search*').as('empSearch')
    cy.get('#emp-search-input').type('Иванова')
    cy.wait('@empSearch')
    cy.get('#emp-results button').first().click()
    cy.get('form[action="/admin/users"]').contains('button', 'Создать').click()
    cy.contains('не менее 3')
  })

  it('ошибка при создании дублирующего логина', () => {
    cy.visit('/admin/users')
    cy.get('form[action="/admin/users"]').within(() => {
      cy.get('input[name=login]').type('admin')
      cy.get('input[name=password]').type('password123')
      cy.get('select[name=role]').select('admin')
    })
    cy.intercept('GET', '/admin/users/employee-search*').as('empSearch')
    cy.get('#emp-search-input').type('Петров')
    cy.wait('@empSearch')
    cy.get('#emp-results button').first().click()
    cy.get('form[action="/admin/users"]').contains('button', 'Создать').click()
    cy.contains('уже существует')
  })

  it('страница сотрудников показывает список с пагинацией', () => {
    cy.visit('/admin/employees')
    cy.contains('Сотрудники')
    cy.get('table tbody tr').should('have.length.greaterThan', 0)
  })

  it('страница отделов показывает руководителей', () => {
    cy.visit('/admin/departments')
    cy.contains('Отделы')
    cy.contains('Engineering')
    cy.contains('Sales')
  })

  it('нельзя изменить собственную роль', () => {
    cy.visit('/admin/users')
    cy.contains('tr', 'admin').within(() => {
      cy.contains('(вы)').should('exist')
      cy.get('form[action*="/role"]').should('not.exist')
    })
  })

  it('страница посещаемости с фильтром по дате', () => {
    cy.visit('/admin/attendance')
    cy.intercept('GET', '/admin/attendance*').as('attFilter')
    cy.get('input[name=from]').type('2026-04-01')
    cy.get('input[name=to]').type('2026-04-30')
    cy.contains('button', 'Найти').click()
    cy.wait('@attFilter')
    cy.get('table tbody tr').should('have.length.greaterThan', 0)
  })

})
