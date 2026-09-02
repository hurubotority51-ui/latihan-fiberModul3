package main

type StudentRepository interface {
	List(q ListQuery) ([]Student, int, error)
	GetByID(id int) (Student, error)
	ExistsByNIM(nim string) (bool, error)
	Create(student CreateStudentRequest) (Student, error)
}
