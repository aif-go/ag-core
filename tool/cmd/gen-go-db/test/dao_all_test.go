package test

import (
	"testing"
)

// TestAllDaoMethods 一次性执行 dao 包所有方法的集成测试（统一测试入口）
//
// 运行方式：
//
//	go test ./test/ -run TestAllDaoMethods -v
//
// 说明：
//   - 覆盖 9 个 DAO 方法（InsertOne / InsertOneIgnoreZeroValCols / UpdateByPrimaryKey /
//     UpdateByPrimaryKeyIngoreZeroValCols / FindByPrimaryKey / FindByStruct /
//     FindByCustomerRule / FindByCondition / FindFirstOneByCondition）
//   - 涉及表：tm_student、tm_teacher、tm_no、tm_no_index、tm_no_primary
//   - 依赖本地 MySQL（localhost:3306/process），需先启动数据库
//   - 各测试自备数据并在结束后清理，不污染既有种子数据
func TestAllDaoMethods(t *testing.T) {
	daoMethodTests := []struct {
		name string
		fn   func(t *testing.T)
	}{
		// ==================== FindByStruct ====================
		{"FindByStruct-学生(复合主键+索引)", TestStudentFindByStruct},
		{"FindByStruct-教师(单主键+自定义规则)", TestTeacherFindByStruct},
		{"FindByStruct-无主键无索引表", TestTmNoFindByStruct},
		{"FindByStruct-有主键无索引表", TestTmNoIndexFindByStruct},
		{"FindByStruct-无主键有索引表", TestTmNoPrimaryFindByStruct},

		// ==================== InsertOne ====================
		{"InsertOne-学生", TestStudentInsertOne},
		{"InsertOne-教师", TestTeacherInsertOne},

		// ==================== InsertOneIgnoreZeroValCols ====================
		{"InsertOneIgnoreZeroValCols-学生", TestStudentInsertOneIgnoreZeroValCols},
		{"InsertOneIgnoreZeroValCols-教师", TestTeacherInsertOneIgnoreZeroValCols},

		// ==================== UpdateByPrimaryKey ====================
		{"UpdateByPrimaryKey-学生", TestStudentUpdateByPrimaryKey},
		{"UpdateByPrimaryKey-教师", TestTeacherUpdateByPrimaryKey},
		{"UpdateByPrimaryKey-无主键表", TestTmNoUpdateByPrimaryKey},

		// ==================== UpdateByPrimaryKeyIngoreZeroValCols ====================
		{"UpdateByPrimaryKeyIngoreZeroValCols-学生", TestStudentUpdateByPrimaryKeyIngoreZeroValCols},
		{"UpdateByPrimaryKeyIngoreZeroValCols-教师", TestTeacherUpdateByPrimaryKeyIngoreZeroValCols},
		{"UpdateByPrimaryKeyIngoreZeroValCols-无主键表", TestTmNoUpdateByPrimaryKeyIngoreZeroValCols},

		// ==================== FindByPrimaryKey ====================
		{"FindByPrimaryKey-学生", TestStudentFindByPrimaryKey},
		{"FindByPrimaryKey-学生-部分主键", TestStudentFindByPrimaryKeyPartial},
		{"FindByPrimaryKey-教师", TestTeacherFindByPrimaryKey},
		{"FindByPrimaryKey-有主键无索引表", TestTmNoIndexFindByPrimaryKey},
		{"FindByPrimaryKey-无主键表", TestTmNoFindByPrimaryKey},

		// ==================== FindByCustomerRule ====================
		{"FindByCustomerRule-参数校验", TestTeacherFindByCustomerRule_Validations},
		{"FindByCustomerRule-FieldMask", TestTeacherFindByCustomerRule_FieldMask},
		{"FindByCustomerRule-FindByNameNadAddress", TestTeacherFindByCustomerRule_FindByNameNadAddress},
		{"FindByCustomerRule-FindByPhone", TestTeacherFindByCustomerRule_FindByPhone},

		// ==================== FindByCondition ====================
		{"FindByCondition-学生", TestStudentFindByCondition},
		{"FindByCondition-教师", TestTeacherFindByCondition},

		// ==================== FindFirstOneByCondition ====================
		{"FindFirstOneByCondition-学生", TestStudentFindFirstOneByCondition},
		{"FindFirstOneByCondition-教师", TestTeacherFindFirstOneByCondition},
	}

	for _, tt := range daoMethodTests {
		t.Run(tt.name, tt.fn)
	}
}
