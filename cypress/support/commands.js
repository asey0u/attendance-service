Cypress.Commands.add('login', (login, password) => {
  cy.visit('/login')
  cy.get('input[name=login]').type(login)
  cy.get('input[name=password]').type(password)
  cy.contains('button', 'Войти').click()
})

Cypress.Commands.add('loginFast', (login, password) => {
  cy.request({
    method: 'POST',
    url: '/login',
    form: true,
    body: { login, password },
    followRedirect: false,
  })
})
