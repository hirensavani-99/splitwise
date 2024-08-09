package models

const (
	//User Query
	QueryToSaveUser = "INSERT INTO users (email, password) VALUES ($1, $2) RETURNING id;"

	//Expense Query
	QueryToPostExpense = `
	INSERT INTO expense (
		description, amount, currency, category, added_at,
		is_recurring, recurring_period, notes, group_id, added_by,add_to,split_type
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,$11,$12
	) RETURNING id;`

	QueryToGetExpense = `
	Select * from expense where group_id=$1
	`
	QueryToGetExpenseByExpenseId = `
	Select * from expense where id=$1;
	`
	QueryToCheckIsExpenseExists = `SELECT EXISTS (SELECT 1 FROM expense WHERE id=$1);`

	QueryToUpdateExpense = `UPDATE expense
	SET `

	QueryToDeleteExpense = `Delete from expense where id=$1;`

	//Group query
	QueryToGetGroupType = `Select simplify_debt from groups where id=$1;`

	QueryToCheckIsMemberOfGroup = `SELECT EXISTS (SELECT 1 FROM group_member WHERE user_id = $1 AND group_id = $2);`

	QueryToSaveGroup = "INSERT INTO groups (name,description,simplify_debt) VALUES ($1, $2,$3) RETURNING id;"

	QueryToAddGroupMember = "INSERT INTO group_member (group_id,user_id) VALUES ($1, $2);"

	QueryToGetGroupsIdByUserId = `SELECT group_id FROM group_member WHERE user_id=$1;`

	//Balance query
	QueryToGetBalances = `	
	SELECT from_user_id, to_user_id,group_id, amount
	FROM Balances
	WHERE from_user_id = $1 OR to_user_id = $1;
`
	QueryToGetExistingBalances = `SELECT from_user_id, to_user_id, amount , group_id FROM BALANCES WHERE ((from_user_id = $1 AND to_user_id = $2) OR (from_user_id = $2 AND to_user_id = $1)) AND group_id=$3;`

	QueryToGetBalanceByGrouId = `Select from_user_id , to_user_id , group_id , amount from balances where group_id=$1;`

	QueryToPostBalances = `
	INSERT INTO BALANCES (from_user_id, to_user_id ,group_id, amount, created_at)
	VALUES ($1, $2, $3, $4,$5);
`

	QueryToUpdateBalances = `UPDATE BALANCES SET amount=$4 WHERE ((from_user_id = $1 AND to_user_id = $2) OR (from_user_id = $2 AND to_user_id = $1)) AND group_id=$3;`

	QueryToUpdateBalancesData = `
	UPDATE BALANCES
	SET amount=$4, from_user_id=$2, to_user_id=$1
	WHERE ((from_user_id = $1 AND to_user_id = $2) OR (from_user_id = $2 AND to_user_id = $1) AND group_id=$3);
`

	QueryToDeleteUnnecessaryBalances = `
	DELETE FROM balances
	WHERE group_id = $1
	AND (from_user_id, to_user_id, group_id, amount) NOT IN `

	//Wallet
	QueryToSaveWallet = `
	INSERT INTO wallets (
		user_id, balance, currency
	) VALUES (
		$1, $2, $3
	);`

	QueryToGetWalletDataByUserId = `
	SELECT user_id, balance, currency, createdAt, updatedAt
	FROM wallets
	WHERE user_id = $1;
`
	QueryToGetWalletBalanceByUserId = `SELECT BALANCE FROM Wallets WHERE USER_ID=$1;`

	QueryToUpdateWalletBalance = `UPDATE Wallets SET BALANCE=$2 , updatedAt=$3 WHERE USER_ID=$1;`

	//comments

	QueryToSaveComments = `INSERT INTO comments (
			expense_id,
			user_id,
			content
		) VALUES (
			$1,$2,$3
		);`

	QueryToGetCommentsByExpenseId = `Select * from comments where expense_id=$1`
)
