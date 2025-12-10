describe("Customer Import & Export", () => {
  it("opens import and export dialogs", () => {
    cy.visit("/customer");
    cy.contains("导入客户").click();
    cy.contains("上传已填写的文件").should("exist");
    cy.contains("取消").click();

    cy.contains("导出客户").click();
    cy.contains("选择导出字段").should("exist");
    cy.contains("取消").click();
  });
});
