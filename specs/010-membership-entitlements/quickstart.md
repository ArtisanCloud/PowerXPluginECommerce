# Quickstart: 订阅会籍权益代币发放

## 前置条件
- 已配置订阅计划 metadata（membershipTierId / benefitIds / tokenCode / tokenAmount）
- 支付回调可正常到达

## 验证步骤
1. 创建订阅订单并完成支付。
2. 触发支付回调后，查询客户权益与代币余额。
3. 验证会籍状态已生效，权益与代币余额已更新。

## 预期结果
- 会籍/权益/代币在 60 秒内可查询到。
- 重复回调不会重复发放。
