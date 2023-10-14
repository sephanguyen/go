package io.manabie.demo.usermgmt.entities;

import java.util.List;

public class Students {
    private List<Student> students;

    public void setStudents(List<Student> students) {
        this.students = students;
    }

    public List<Student> getStudents() {
        return this.students;
    }
}
