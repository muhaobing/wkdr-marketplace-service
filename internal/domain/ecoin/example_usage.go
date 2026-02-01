package ecoin

// 这个文件展示了如何使用 ecoin 积分模块的示例代码
// 注意：这只是示例，实际使用时请根据业务需求进行调整

import (
	"context"
	"log"

	"wdkr-marketplace-service/internal/domain/ecoin/ecoin_model"
	"wdkr-marketplace-service/internal/domain/ecoin/repo"
)

// ExampleUsage 积分模块使用示例
func ExampleUsage(ctx context.Context) {
	// 方式1：直接创建
	ecoinRepo := repo.NewEcoinRepo()
	ecoinService := NewEcoinService(ecoinRepo)

	// 方式2：使用工厂模式（推荐）
	serviceFactory := NewServiceFactory()
	ecoinService = serviceFactory.GetEcoinService()

	// 示例用户ID
	userId := uint64(12345)

	// 1. 获取用户积分信息（如果不存在会自动初始化）
	userEcoin, err := ecoinService.GetUserEcoin(ctx, userId)
	if err != nil {
		log.Printf("获取用户积分失败: %v", err)
		return
	}
	log.Printf("用户 %d 当前积分: %.2f", userId, userEcoin.AvailableStock)

	// 2. 为用户增加积分（比如订单奖励）
	addReq := &AddEcoinRequest{
		UserId:      userId,
		Amount:      100.0,
		SourceType:  ecoin_model.SourceTypeOrder,
		SourceId:    "ORDER_123456",
		Description: "订单购买奖励积分",
	}
	transaction, err := ecoinService.AddEcoin(ctx, addReq)
	if err != nil {
		log.Printf("增加积分失败: %v", err)
		return
	}
	log.Printf("积分增加成功，流水ID: %d，当前余额: %.2f", transaction.Id, transaction.AfterStock)

	// 3. 扣除积分（比如积分消费）
	deductReq := &DeductEcoinRequest{
		UserId:      userId,
		Amount:      50.0,
		SourceType:  ecoin_model.SourceTypeConsume,
		SourceId:    "CONSUME_789",
		Description: "积分兑换商品",
	}
	transaction, err = ecoinService.DeductEcoin(ctx, deductReq)
	if err != nil {
		log.Printf("扣除积分失败: %v", err)
		return
	}
	log.Printf("积分扣除成功，流水ID: %d，当前余额: %.2f", transaction.Id, transaction.AfterStock)

	// 4. 查询积分流水记录
	listReq := &EcoinTransactionListRequest{
		UserId: userId,
		Offset: 0,
		Limit:  10,
	}
	transactions, err := ecoinService.GetEcoinTransactionList(ctx, listReq)
	if err != nil {
		log.Printf("获取积分流水失败: %v", err)
		return
	}
	log.Printf("用户 %d 的积分流水记录数量: %d", userId, len(transactions))

	// 5. 查询特定的积分流水记录
	if len(transactions) > 0 {
		transactionId := transactions[0].Id
		singleTransaction, err := ecoinService.GetEcoinTransaction(ctx, transactionId)
		if err != nil {
			log.Printf("获取积分流水详情失败: %v", err)
			return
		}
		log.Printf("流水详情 - ID: %d, 金额: %.2f, 类型: %d, 来源: %s",
			singleTransaction.Id,
			singleTransaction.Amount,
			singleTransaction.TxType,
			singleTransaction.SourceType)
	}
}

/*
依赖注入架构说明：

新的架构设计采用依赖注入模式，具有以下优势：
1. 解耦合：Service层不直接依赖数据库连接，而是依赖Repo接口
2. 可测试性：可以轻松为Service层注入Mock的Repo进行单元测试
3. 灵活性：可以在不同场景下注入不同的Repo实现
4. 职责分离：数据库连接的获取逻辑完全封装在Repo层

初始化方式：
- 方式1：手动创建repo和service实例
- 方式2：使用ServiceFactory工厂模式（推荐）

常用的积分操作类型说明：

1. 积分来源类型 (SourceType)：
   - ecoin_model.SourceTypeSystem: "system" - 系统赠送
   - ecoin_model.SourceTypeOrder: "order" - 订单奖励
   - ecoin_model.SourceTypeConsume: "consume" - 积分消费
   - ecoin_model.SourceTypeRefund: "refund" - 退款返还
   - ecoin_model.SourceTypeTransfer: "transfer" - 转账

2. 交易类型 (TxType)：
   - ecoin_model.TransactionTypeAdd (1): 增加积分
   - ecoin_model.TransactionTypeDeduct (2): 扣除积分

3. 交易状态 (Status)：
   - ecoin_model.TransactionStatusPending (0): 处理中
   - ecoin_model.TransactionStatusCompleted (1): 已完成
   - ecoin_model.TransactionStatusFailed (2): 已失败

数据库表结构说明：

1. user_ecoin 表：维护用户积分余额
   - id: 主键ID
   - user_id: 用户ID（唯一索引）
   - available_stock: 可用积分余额
   - ctime, mtime: 创建和修改时间

2. ecoin_transaction 表：积分变动流水记录
   - id: 主键ID
   - user_id: 用户ID
   - amount: 积分变动数量（正数表示增加，负数表示扣除）
   - before_stock: 操作前积分余额
   - after_stock: 操作后积分余额
   - tx_type: 交易类型
   - source_type: 来源类型
   - source_id: 来源业务ID
   - description: 描述信息
   - status: 交易状态
   - ctime, mtime: 创建和修改时间
*/
