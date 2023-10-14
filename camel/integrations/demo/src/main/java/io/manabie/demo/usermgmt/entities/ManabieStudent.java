package io.manabie.demo.usermgmt.entities;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;

public class ManabieStudent {

    private Parent toParent(HashMap csvData) {
        Parent parent = new Parent();
        parent.setParentNumber(csvData.get("保護者番号").toString());
        parent.setParentName(csvData.get("保護者氏名").toString());
        parent.setParentRawPassword(csvData.get("保護者パスワード").toString());
        parent.setParentEmail(csvData.get("保護者メール").toString());
        return parent;
    }

    private Student toStudent(HashMap csvData) {
        Student student = new Student();
        student.setParent(toParent(csvData));
        student.setCustomerNumber(csvData.get("顧客番号").toString());
        student.setStudentNumber(csvData.get("生徒番号").toString());

        student.setName(csvData.get("氏名（ニックネーム）").toString());
        student.setPasswordRaw(csvData.get("パスワード").toString());
        student.setStudentEmail(csvData.get("生徒メール").toString());
        student.setLocations(csvData.get("G1（所属）").toString());
        student.setTagG2(csvData.get("G2（セグメント）").toString());
        student.setTagG3(csvData.get("G3（生徒区分）").toString());
        student.setTagG4(csvData.get("G4（本校）").toString());
        student.setTagG5(csvData.get("G5（学年）").toString());
        student.setCourses(csvData.get("所持商品").toString());
        student.setDeleteFlag(csvData.get("削除フラグ").toString());
        student.setGraduationExpectedDate("");
        student.setUserIDAttr("");
        student.setEmailDomain("");
        return student;
    }

    public Students exchange(HashMap csvData) {
        List<Student> list = new ArrayList<Student>();
        list.add(toStudent(csvData));

        Students students = new Students();
        students.setStudents(list);

        return students;
    }

}
