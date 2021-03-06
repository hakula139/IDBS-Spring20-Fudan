package main

import (
	"fmt"
	"time"

	// YOUR CODE BEGIN remove the follow packages if you don't need them
	"sync"
	"reflect"
	// YOUR CODE END

	_ "github.com/go-sql-driver/mysql"
	sql "github.com/jmoiron/sqlx"
)

var (
	// YOUR CODE BELOW
	EvaluatorID   = "18307130031" // your student id, e.g. 18307130177
	SubmissionDir = "../../../ass1/submission" // the relative path the the submission directory of assignment 1, it should be "../../../ass1/submission/"
	User          = "root" // the user name to connect the database, e.g. root
	Password      = "xudinghuan" // the password for the user name, e.g. xxx
	// YOUR CODE END
)

// ConcurrentCompareAndInsert is similar with compareAndInsert in `main.go`, but it is concurrent and faster!
func ConcurrentCompareAndInsert(subs map[string]*Submission) {
	start := time.Now()
	defer func() {
		db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/ass1_result_evaluated_by_%s", User, Password, EvaluatorID))
		if err != nil {
			panic(nil)
		}
		rows, err := db.Query("SELECT COUNT(*) FROM comparison_result")
		if err != nil {
			panic(err)
		}
		rows.Next()
		var cnt int
		err = rows.Scan(&cnt)
		if err != nil {
			panic(err)
		}
		if cnt == 0 {
			panic("ConcurrentCompareAndInsert Not Implemented")
		}
		fmt.Println("ConcurrentCompareAndInsert takes ", time.Since(start))
	}()
	// YOUR CODE BEGIN
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/ass1_result_evaluated_by_%s", User, Password, EvaluatorID))
	if err != nil{
		fmt.Println("fail to connect database")
		panic(err)	
	}
	
	stmt, err := db.Prepare("INSERT INTO comparison_result VALUES (?,?,?,?)")
	if err != nil{
		fmt.Println("fail to prepare statement")
		panic(err)
	}


	var wg sync.WaitGroup
	const num = 35
	
	type result struct{
		submitter string
		comparer  string
		item 	  int
		is_equal  int 
	}

	ch := make(chan result, num)

	job := func(){
		tx, err := db.Begin()
		if err != nil{
			fmt.Println("fail to initialize transaction")
			panic(err)
		}

		exec := tx.Stmt(stmt)

		defer func(){
			err := tx.Commit()
			if err != nil{
				fmt.Println("fail to commit transaction")
				panic(err)
			}
		}()
		
		for i := range ch{
			_, err := exec.Exec(i.submitter, i.comparer, i.item, i.is_equal)
			if err != nil{
				fmt.Println("fail to perform insertions")	
				panic(err)			
			}
		}

		wg.Done()
		exec.Close()
	}
	
	for i := 0; i < num; i++{
		wg.Add(1)
		go job()
	}

	for submitter, sub := range subs {
		for comparer, sub2 := range subs {
			for i := 0; i < NumSQL; i++ {
				var equal int
				if reflect.DeepEqual(sub.sqlResults[i], sub2.sqlResults[i]) {
					equal = 1
				} else {
					equal = 0
				}
				
				ch <- result{submitter, comparer, i + 1, equal}
			}
		}
	}
	close(ch)

	wg.Wait()
	
	// YOUR CODE END
}
 
// GetScoreSQL returns a string which contains only ONE SQL to be executed, which collects the data in table
// `comparision_result` and inserts the score of each submitter on each query into table `score`
func GetScoreSQL() string {
	var SQL string
	SQL = "SELECT 1" // ignore this line, it just makes the returned SQL a valid SQL if you haven't written yours.
	// YOUR CODE BEGIN
	SQL = `INSERT INTO score 
	       WITH onenum AS(SELECT submitter, item, COUNT(comparer) AS vote 
	       FROM comparison_result
	       WHERE is_equal = '1' 
	       GROUP BY submitter, item), 
	       standard AS(SELECT item, MAX(vote) AS highest 
	       FROM onenum GROUP BY item) 
	       SELECT submitter, onenum.item, IF(vote = highest, 1, 0) AS score, vote 
	       FROM onenum, standard 
	       WHERE onenum.item = standard.item
	       ORDER BY onenum.item`


	// YOUR CODE END
	return SQL
}

func GetScore(db *sql.DB, subs map[string]*Submission) {
	// YOUR CODE BEGIN
	rows, err := db.Query("SELECT submitter, item, score FROM score")
	if err != nil{
		panic(err)
	}

	for rows.Next() {
		var submitter string
		var item, score int
		rows.Scan(&submitter, &item, &score)
		subs[submitter].score[item] = score
	}

	// YOUR CODE END
}
