import './commands'

beforeEach(() => {
  cy.intercept('GET', 'https://cdn.tailwindcss.com**', { body: '' })
  cy.intercept('GET', 'https://fonts.googleapis.com**', { body: '' })
  cy.intercept('GET', 'https://fonts.gstatic.com**',    { body: '' })
})
