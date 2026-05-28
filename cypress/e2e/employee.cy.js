describe('Сотрудник', () => {

  beforeEach(() => {
    cy.loginFast('user1', 'employee1')
    cy.visit('/me')
  })

  it('дашборд загружается и показывает рабочий день', () => {
    cy.contains('Главная')
    cy.contains('Сегодня')
    cy.contains('Статистика')
  })

  it('кнопка прихода или ухода присутствует', () => {
    cy.get('body').then($body => {
      const hasCheckIn  = $body.find('form[action="/me/check-in"]').length > 0
      const hasCheckOut = $body.find('form[action="/me/check-out"]').length > 0
      const dayDone     = $body.text().includes('завершён')
      expect(hasCheckIn || hasCheckOut || dayDone).to.be.true
    })
  })

  it('отметка прихода — если ещё не отмечен', () => {
    cy.get('body').then($body => {
      if ($body.find('form[action="/me/check-in"]').length > 0) {
        cy.get('form[action="/me/check-in"] button').click()
        cy.url().should('include', '/me')
        cy.contains('Отметить уход')
      }
    })
  })

  it('отметка ухода — если приход уже есть', () => {
    cy.get('body').then($body => {
      if ($body.find('form[action="/me/check-out"]').length > 0) {
        cy.get('form[action="/me/check-out"] button').click()
        cy.url().should('include', '/me')
        cy.contains('завершён')
      }
    })
  })

  it('страница посещаемости показывает историю', () => {
    cy.visit('/me/attendance')
    cy.contains('Моя история')
    cy.contains('Дней присутствия')
  })

  it('создание заявки', () => {
    cy.visit('/me/tickets')
    cy.get('input[name=ticket_date]').type('2026-03-01')
    cy.get('textarea[name=reason]').type('Тестовая заявка из Cypress')
    cy.contains('button', 'Отправить').click()
    cy.url().should('include', '/me/tickets')
    cy.contains('Тестовая заявка из Cypress')
  })

  it('удаление заявки в статусе pending', () => {
    cy.visit('/me/tickets')
    cy.intercept('DELETE', '/me/tickets/*').as('deleteTicket')
    cy.on('window:confirm', () => true)
    cy.get('table tbody tr')
      .filter(':has(button:contains("Удалить"))')
      .contains('Тестовая заявка из Cypress')
      .closest('tr')
      .then($row => {
        cy.wrap($row).find('button').click()
        cy.wait('@deleteTicket').its('response.statusCode').should('eq', 200)
        cy.wrap($row).should('not.exist')
      })
  })

})
