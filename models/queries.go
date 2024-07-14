package models

const (

	//Expense Query
	QueryToPostExpense = `
	INSERT INTO expense (
		description, amount, currency, category, added_at,
		is_recurring, recurring_period, notes, group_id, added_by
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
	) RETURNING id;`

	//Group query
	QueryToGetGroupType = `Select simplify_debt from groups where id=$1;`

	//Balance query

	QueryToGetExistingBalances = `SELECT from_user_id, to_user_id, amount , group_id FROM BALANCES WHERE ((from_user_id = $1 AND to_user_id = $2) OR (from_user_id = $2 AND to_user_id = $1)) AND group_id=$3;`

	QueryToPostBalances = `
	INSERT INTO BALANCES (from_user_id, to_user_id ,group_id, amount, created_at)
	VALUES ($1, $2, $3, $4,$5)
`

	QueryToUpdateBalances = `UPDATE BALANCES SET amount=$4 WHERE ((from_user_id = $1 AND to_user_id = $2) OR (from_user_id = $2 AND to_user_id = $1)) AND group_id=$3;`

	QueryToUpdateBalancesData = `
	UPDATE BALANCES
	SET amount=$4, from_user_id=$2, to_user_id=$1
	WHERE ((from_user_id = $1 AND to_user_id = $2) OR (from_user_id = $2 AND to_user_id = $1) AND group_id=$3);
`

	
)
