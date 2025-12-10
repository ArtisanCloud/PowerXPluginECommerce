describe("Customer Directory", () => {
  it("renders filters and table shell", () => {
    cy.visit("/customer");
    cy.contains("客户运营").should("be.visible");
    cy.get("input[placeholder*='姓名']").should("exist");
    cy.contains("批量").should("exist");
  });
});
