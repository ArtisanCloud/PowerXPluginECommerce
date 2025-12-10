describe("Customer CRUD affordances", () => {
  beforeEach(() => {
    cy.intercept("GET", "**/admin/customers", { fixture: "customers/list.json" }).as("listCustomers");
  });

  it("opens create modal and submits payload", () => {
    cy.visit("/customer");
    cy.wait(["@listCustomers"]);
    cy.contains("添加客户").click();
    cy.get('input[placeholder="输入客户姓名..."]').type("自动化测试用户");
    cy.get('input[type="tel"]').type("13800001111");
    cy.get('input[type="email"]').type("crud@example.com");
    cy.intercept("POST", "**/admin/customers", {
      id: "cust-e2e",
      name: "自动化测试用户",
      type: "individual",
    }).as("createCustomer");
    cy.contains("创建客户").click();
    cy.wait("@createCustomer").its("request.body").should((body) => {
      expect(body.name).to.eq("自动化测试用户");
      expect(body.phone).to.contain("138");
    });
    cy.contains("客户创建成功").should("be.visible");
  });

});
