describe('Аутентификация', () => {

  it('успешный вход как admin — редирект на /admin', () => {
    cy.login('admin', 'admin')
    cy.url().should('include', '/admin')
    cy.contains('Обзор')
  })

  it('успешный вход как manager — редирект на /manager', () => {
    cy.loginFast('manager', 'manager123')
    cy.visit('/manager')
    cy.url().should('include', '/manager')
    cy.contains('Моя команда')
  })

  it('успешный вход как employee — редирект на /me', () => {
    cy.login('user1', 'employee1')
    cy.url().should('include', '/me')
  })

  it('неверный пароль — показывает ошибку', () => {
    cy.visit('/login')
    cy.get('input[name=login]').type('admin')
    cy.get('input[name=password]').type('wrongpassword')
    cy.contains('button', 'Войти').click()
    cy.url().should('include', '/login')
    cy.contains('Неверный логин или пароль')
  })

  it('выход из системы — редирект на /login', () => {
    cy.login('admin', 'admin')
    cy.url().should('include', '/admin')
    cy.get('form[action="/logout"] button').first().click()
    cy.url().should('include', '/login')
  })

  it('без авторизации /admin редиректит на /login', () => {
    cy.visit('/admin')
    cy.url().should('include', '/login')
  })

})
