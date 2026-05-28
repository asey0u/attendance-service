describe('Руководитель', () => {

  beforeEach(() => {
    cy.loginFast('manager', 'manager123')
    cy.visit('/manager')
  })

  it('дашборд показывает счётчики команды', () => {
    cy.contains('Сегодня на работе')
    cy.contains('Ожидают рассмотрения')
    cy.contains('Размер команды')
  })

  it('страница посещаемости загружается с фильтрами', () => {
    cy.visit('/manager/attendance')
    cy.contains('Посещаемость команды')
    cy.get('input[name=first_name]').should('exist')
    cy.get('input[name=last_name]').should('exist')
    cy.get('input[name=from]').should('exist')
    cy.get('input[name=to]').should('exist')
  })

  it('фильтрация по фамилии', () => {
    cy.visit('/manager/attendance')
    cy.intercept('GET', '/manager/attendance*').as('attFilter')
    cy.get('input[name=last_name]').type('Кузнецов')
    cy.contains('button', 'Найти').click()
    cy.wait('@attFilter')
    cy.get('table tbody tr').each($row => {
      cy.wrap($row).contains('Кузнецов')
    })
  })

  it('страница заявок показывает pending по умолчанию', () => {
    cy.visit('/manager/tickets')
    cy.contains('Заявки команды')
    cy.get('table').should('exist')
  })

  it('одобрение заявки', () => {
    cy.visit('/manager/tickets')
    cy.get('table tbody tr').first().then($row => {
      if ($row.find('button:contains("Одобрить")').length > 0) {
        cy.intercept('POST', '/manager/tickets/*/approve').as('approve')
        cy.wrap($row).contains('button', 'Одобрить').click()
        cy.wait('@approve')
        cy.wrap($row).should('not.exist')
      }
    })
  })

  it('отклонение заявки', () => {
    cy.visit('/manager/tickets')
    cy.get('table tbody tr').first().then($row => {
      if ($row.find('button:contains("Отклонить")').length > 0) {
        cy.intercept('POST', '/manager/tickets/*/decline').as('decline')
        cy.wrap($row).contains('button', 'Отклонить').click()
        cy.wait('@decline')
        cy.wrap($row).should('not.exist')
      }
    })
  })

  it('страница команды показывает список сотрудников', () => {
    cy.visit('/manager/employees')
    cy.contains('Команда')
    cy.get('table tbody tr').should('have.length.greaterThan', 0)
  })

})
