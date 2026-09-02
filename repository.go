package main

type StudentRepository interface {
	FindAll(q ListQuery) ([]Student, int, error)
	FindByID(id int) (Student, error)
	Create(student CreateStudentRequest) (Student, error)
	Update(id int, student ReplaceStudentRequest) (Student, error)
	Delete(id int) error
	ExistsByNIM(nim string) (bool, error)
}
