# Commission Calculation System

## Overview

The Commission Calculation System provides a comprehensive framework for calculating, tracking, and managing sales agent commissions with support for tiered rates, product-specific bonuses, and automated approval workflows.

## Features

### 1. Base Commission Calculation
- Configurable base commission rate per agent
- Applied to order subtotal (excluding shipping/taxes)
- Automatic calculation on order completion

### 2. Tiered Commission Structure
- Volume-based commission tiers
- Higher rates for larger orders
- Unlimited tier support
- Per-agent tier configuration

### 3. Product-Specific Bonuses
- Additional commission for specific products
- Promotional product incentives
- Time-limited product bonuses
- Agent-specific or all-agent application

### 4. Category Bonuses
- Extra commission for product categories
- Category-wide promotions
- Strategic category focus incentives

### 5. Team Bonuses
- Team-based commission boost
- Team performance incentives
- Shared success rewards

### 6. Commission Status Workflow
- **Pending**: Awaiting approval
- **Approved**: Approved for payment
- **Paid**: Payment completed
- **Rejected**: Not eligible for payment

---

## Commission Calculation Formula

### Base Calculation
```
Commission = Order Subtotal × (Commission Rate / 100)
```

### With Tiered Rates
```
1. Determine order total
2. Find matching tier (by amount range)
3. Apply tier rate instead of base rate
4. Commission = Order Subtotal × (Tier Rate / 100)
```

### With Product Bonuses
```
Base Commission = Subtotal × (Base Rate / 100)
Product Bonus = Subtotal × (Product Bonus Rate / 100)
Total Commission = Base Commission + Product Bonus
```

### With Team Boost
```
Effective Rate = Base Rate + Team Boost
Commission = Order Subtotal × (Effective Rate / 100)
```

### Complete Formula
```
Effective Rate = (Tier Rate OR Base Rate) + Team Boost
Base Commission = Subtotal × (Effective Rate / 100)
Product Bonuses = ΣProduct (Subtotal × Product Bonus Rate / 100)
Category Bonuses = ΣCategory (Subtotal × Category Bonus Rate / 100)

Total Commission = Base Commission + Product Bonuses + Category Bonuses
```

---

## Database Schema

### agent_commissions Table

```sql
CREATE TABLE sales.agent_commissions (
    id UUID PRIMARY KEY,
    agent_id UUID NOT NULL,
    order_id UUID NOT NULL,
    order_total DECIMAL(12,2),
    commission_rate DECIMAL(5,2),
    commission_amount DECIMAL(10,2),
    based_on_amount DECIMAL(12,2),
    status VARCHAR(20) DEFAULT 'pending',
    notes TEXT,
    approved_by UUID,
    approved_at TIMESTAMP,
    paid_at TIMESTAMP,
    rejected_at TIMESTAMP,
    rejection_reason TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

### commission_tiers Table

```sql
CREATE TABLE sales.commission_tiers (
    id UUID PRIMARY KEY,
    agent_id UUID NOT NULL,
    min_amount DECIMAL(12,2) NOT NULL,
    max_amount DECIMAL(12,2), -- NULL for unlimited
    rate DECIMAL(5,2) NOT NULL,
    description VARCHAR(255),
    is_active BOOLEAN DEFAULT TRUE
);
```

### product_commission_rates Table

```sql
CREATE TABLE sales.product_commission_rates (
    id UUID PRIMARY KEY,
    product_id UUID NOT NULL,
    product_name VARCHAR(255),
    commission_bonus_rate DECIMAL(5,2) DEFAULT 0,
    applies_to_all_agents BOOLEAN DEFAULT FALSE,
    agent_ids UUID[],
    start_date TIMESTAMP,
    end_date TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE
);
```

### category_commission_rates Table

```sql
CREATE TABLE sales.category_commission_rates (
    id UUID PRIMARY KEY,
    category_id UUID NOT NULL,
    category_name VARCHAR(255),
    commission_bonus_rate DECIMAL(5,2) DEFAULT 0,
    applies_to_all_agents BOOLEAN DEFAULT FALSE,
    agent_ids UUID[],
    start_date TIMESTAMP,
    end_date TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE
);
```

---

## API Usage

### Calculate Commission

```go
calculator := services.NewCommissionCalculatorService(db, logger)

request := &services.CommissionCalculationRequest{
    OrderID:        orderUUID,
    AgentID:        agentUUID,
    OrderTotal:     1500.00,
    OrderSubtotal:  1400.00,
    ShippingCost:   100.00,
    DiscountAmount: 50.00,
    ProductIDs:     []uuid.UUID{product1UUID, product2UUID},
    CategoryIDs:    []uuid.UUID{categoryUUID},
}

result, err := calculator.CalculateCommission(request)
if err != nil {
    // Handle error
}

// result.CommissionAmount - Total commission
// result.CommissionRate - Applied rate
// result.TierApplied - Tier used (if any)
// result.Breakdown - Detailed breakdown
```

### Create Commission Record

```go
err := calculator.CreateCommission(result)
if err != nil {
    // Handle error
}
```

### Approve Commission

```go
err := calculator.ApproveCommission(
    commissionID,
    approverID,
    "Approved after order delivery confirmation",
)
```

### Reject Commission

```go
err := calculator.RejectCommission(
    commissionID,
    approverID,
    "Order was cancelled by customer",
)
```

### Mark as Paid

```go
err := calculator.MarkCommissionPaid(
    commissionID,
    time.Now(),
)
```

---

## Examples

### Example 1: Basic Commission (5%)

**Configuration**:
- Agent base rate: 5%
- Order subtotal: RM1,000

**Calculation**:
```
Commission = RM1,000 × 5% = RM50.00
```

---

### Example 2: Tiered Commission

**Configuration**:
- Tier 1: RM0-1000 = 5%
- Tier 2: RM1001-5000 = 7.5%
- Tier 3: RM5001+ = 10%

**Order: RM3,500**
```
Tier Applied: Tier 2 (7.5%)
Commission = RM3,500 × 7.5% = RM262.50
```

**Order: RM6,000**
```
Tier Applied: Tier 3 (10%)
Commission = RM6,000 × 10% = RM600.00
```

---

### Example 3: Product Bonus

**Configuration**:
- Base rate: 5%
- Product "Premium Batik" bonus: +3%
- Order subtotal: RM2,000 (includes Premium Batik)

**Calculation**:
```
Base Commission = RM2,000 × 5% = RM100.00
Product Bonus = RM2,000 × 3% = RM60.00
Total Commission = RM160.00
```

---

### Example 4: Team Boost

**Configuration**:
- Agent base rate: 5%
- Team boost: +2%
- Order subtotal: RM1,500

**Calculation**:
```
Effective Rate = 5% + 2% = 7%
Commission = RM1,500 × 7% = RM105.00
```

---

### Example 5: Complete Calculation

**Configuration**:
- Tier rate (RM3,000 order): 7.5%
- Team boost: +2%
- Category "Silk Batik" bonus: +3%

**Calculation**:
```
Effective Rate = 7.5% + 2% = 9.5%
Base Commission = RM3,000 × 9.5% = RM285.00
Category Bonus = RM3,000 × 3% = RM90.00
Total Commission = RM375.00
```

---

## Commission Workflow

### On Order Creation

1. Order placed by customer
2. Order status: Pending
3. Commission status: Not created yet

### On Order Payment Confirmed

1. Payment verified
2. Commission calculated automatically
3. Commission status: Pending
4. Agent notified

### Commission Approval

1. Manager reviews commission
2. Checks order status (delivered/completed)
3. Approves or rejects
4. Agent notified of decision

### Payment Processing

1. Approved commissions collected
2. Batch payment processed
3. Commissions marked as Paid
4. Agent receives payment
5. Agent's total_earned updated

---

## Approval Workflow

### Auto-Approval Conditions

Can be configured to auto-approve if:
- Order delivered successfully
- No returns/refunds within X days
- Order value above minimum threshold
- Agent has good standing

### Manual Approval Required

- High-value orders (> RM10,000)
- First-time agent orders
- Orders with returns/refunds
- Disputed orders

---

## Commission Status Transitions

```
                    ┌─────────────┐
                    │   PENDING   │
                    └──────┬──────┘
                           │
              ┌────────────┼────────────┐
              │                         │
              ▼                         ▼
       ┌──────────┐             ┌──────────┐
       │ APPROVED │             │ REJECTED │
       └────┬─────┘             └──────────┘
            │
            ▼
       ┌────────┐
       │  PAID  │
       └────────┘
```

**Valid Transitions**:
- Pending → Approved
- Pending → Rejected
- Approved → Paid

**Invalid Transitions**:
- Paid → Any other status (final)
- Rejected → Any other status (final)
- Approved → Rejected

---

## Database Functions

### calculate_agent_commission()

```sql
SELECT sales.calculate_agent_commission(
    'agent-uuid',
    1500.00 -- order total
);
-- Returns: commission amount
```

**Features**:
- Considers agent's base rate
- Applies tiered rates if enabled
- Adds team boost if applicable
- Returns calculated amount

---

### get_agent_commission_stats()

```sql
SELECT * FROM sales.get_agent_commission_stats(
    'agent-uuid',
    '2025-01-01'::timestamp, -- start date (optional)
    '2025-12-31'::timestamp  -- end date (optional)
);
```

**Returns**:
- total_pending
- total_approved
- total_paid
- total_orders
- total_sales

---

## Reporting & Analytics

### Agent Performance Report

```go
perf, err := calculator.CalculateMonthlyPerformance(
    agentID,
    2025, // year
    1,    // month
)

// perf.TotalSales
// perf.TotalOrders
// perf.TotalCommission
// perf.AverageOrderValue
// perf.AverageCommission
```

### Commission Statistics

```go
stats, err := calculator.GetCommissionStats(agentID)

// stats.TotalPending
// stats.TotalApproved
// stats.TotalPaid
// stats.TotalEarned
// stats.TotalUnpaid
// stats.TotalCount
```

---

## Integration with Order Service

### On Order Completion

```go
// In order service, after order is marked as delivered
if order.AgentID != nil {
    req := &services.CommissionCalculationRequest{
        OrderID:        order.ID,
        AgentID:        *order.AgentID,
        OrderTotal:     order.Total,
        OrderSubtotal:  order.Subtotal,
        ShippingCost:   order.ShippingCost,
        DiscountAmount: order.Discount,
        ProductIDs:     extractProductIDs(order.Items),
        CategoryIDs:    extractCategoryIDs(order.Items),
    }

    result, err := commissionCalculator.CalculateCommission(req)
    if err != nil {
        logger.Error("Failed to calculate commission", zap.Error(err))
        return
    }

    result.OrderTotal = order.Total
    if err := commissionCalculator.CreateCommission(result); err != nil {
        logger.Error("Failed to create commission", zap.Error(err))
    }
}
```

### On Order Cancellation

```go
// Find and reject commission if exists
commissions, _ := commissionCalculator.GetCommissionsByOrder(orderID)
for _, comm := range commissions {
    if comm.Status == CommissionStatusPending {
        calculator.RejectCommission(
            comm.ID,
            systemUserID,
            "Order cancelled by customer",
        )
    }
}
```

---

## Best Practices

### 1. Commission Timing

- Calculate after order delivery
- Wait for return period to expire
- Consider payment confirmation
- Handle cancellations gracefully

### 2. Rate Management

- Review rates quarterly
- Adjust based on performance
- Communicate changes in advance
- Document rate change history

### 3. Approval Process

- Set clear approval criteria
- Define approval thresholds
- Automate where possible
- Track approval times

### 4. Payment Schedule

- Regular payment cycles (monthly)
- Batch payment processing
- Clear payment statements
- Maintain payment records

### 5. Dispute Resolution

- Clear dispute process
- Documented decision criteria
- Fair appeal mechanism
- Timely resolution

---

## Monitoring

### Key Metrics

1. **Average Commission Rate**: Mean % across all orders
2. **Total Commissions Paid**: Monthly payout amounts
3. **Pending Commission Value**: Unpaid commissions
4. **Approval Time**: Time from pending to approved
5. **Rejection Rate**: % commissions rejected

### Monitoring Queries

```sql
-- Monthly commission summary
SELECT
    DATE_TRUNC('month', created_at) as month,
    COUNT(*) as total_commissions,
    SUM(commission_amount) as total_amount,
    AVG(commission_amount) as avg_amount
FROM sales.agent_commissions
WHERE created_at >= NOW() - INTERVAL '12 months'
GROUP BY month
ORDER BY month DESC;

-- Top performing agents
SELECT
    a.name,
    COUNT(*) as orders,
    SUM(ac.order_total) as total_sales,
    SUM(ac.commission_amount) as total_commission
FROM sales.agent_commissions ac
JOIN sales.agents a ON a.id = ac.agent_id
WHERE ac.created_at >= NOW() - INTERVAL '30 days'
GROUP BY a.id, a.name
ORDER BY total_commission DESC
LIMIT 10;
```

---

## Future Enhancements

1. **Recurring Commissions**: Subscription-based commissions
2. **Split Commissions**: Multiple agents per order
3. **Commission Clawback**: Reverse on returns
4. **Performance Bonuses**: Target-based bonuses
5. **Commission Advances**: Early payment options
6. **Automated Reconciliation**: Bank integration
7. **Tax Reporting**: Automated tax documents
8. **Mobile App**: Agent commission tracking app

---

**Last Updated**: 2025-12-09
**Version**: 1.0.0
