describe("Membership View", () => {
  it("renders filters and reminder entry", () => {
    cy.visit("/customer/members");
    cy.contains("会员运营视角").should("be.visible");
    cy.contains("批量提醒").should("exist");
    cy.contains("成长值区间").should("exist");
  });
});
