package database

import "testing"

func createTestUsers(t *testing.T, db *Database) (studentID, teacherID int64) {
    userRepo := NewUserRepository(db.DB)
    tID, _ := userRepo.Create(&User{Username: "t", Password: "1", Role: "teacher"})
    sID, _ := userRepo.Create(&User{Username: "s", Password: "1", Role: "student"})
    return sID, tID
}
