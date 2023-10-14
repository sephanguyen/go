package io.manabie.demo.usermgmt.entities;

public class Student extends DomainUser {
    private String UserIDAttr;
    private Parent Parent;
    private String CustomerNumber;
    private String StudentNumber;
    private String Name;
    private String PasswordRaw;
    private String StudentEmail;
    private String Locations;
    private String TagG2;
    private String TagG3;
    private String TagG4;
    private String TagG5;
    private String Courses;
    private String DeleteFlag;
    private String EmailDomain;
    private String GraduationExpectedDate;

    public void setEmailDomain(String EmailDomain) {
        this.EmailDomain = EmailDomain;
    }

    public String getEmailDomain() {
        return this.EmailDomain;
    }

    public void setUserIDAttr(String UserIDAttr) {
        this.UserIDAttr = UserIDAttr;
    }

    public String getUserIDAttr() {
        return this.UserIDAttr;
    }

    public void setGraduationExpectedDate(String GraduationExpectedDate) {
        this.GraduationExpectedDate = GraduationExpectedDate;
    }

    public String getGraduationExpectedDate() {
        return this.GraduationExpectedDate;
    }

    public void setParent(Parent Parent) {
        this.Parent = Parent;
    }

    public Parent getParent() {
        return this.Parent;
    }

    public String getCustomerNumber() {
        return this.CustomerNumber;
    }

    public void setCustomerNumber(String CustomerNumber) {
        this.CustomerNumber = CustomerNumber;
    }

    public String getStudentNumber() {
        return this.StudentNumber;
    }

    public void setStudentNumber(String StudentNumber) {
        this.StudentNumber = StudentNumber;
    }

    public String getName() {
        return this.Name;
    }

    public void setName(String Name) {
        this.Name = Name;
    }

    public String getPasswordRaw() {
        return this.PasswordRaw;
    }

    public void setPasswordRaw(String PasswordRaw) {
        this.PasswordRaw = PasswordRaw;
    }

    public String getStudentEmail() {
        return this.StudentEmail;
    }

    public void setStudentEmail(String StudentEmail) {
        this.StudentEmail = StudentEmail;
    }

    public String getLocations() {
        return this.Locations;
    }

    public void setLocations(String Locations) {
        this.Locations = Locations;
    }

    public String getTagG2() {
        return this.TagG2;
    }

    public void setTagG2(String TagG2) {
        this.TagG2 = TagG2;
    }

    public String getTagG3() {
        return this.TagG3;
    }

    public void setTagG3(String TagG3) {
        this.TagG3 = TagG3;
    }

    public String getTagG4() {
        return this.TagG4;
    }

    public void setTagG4(String TagG4) {
        this.TagG4 = TagG4;
    }

    public String getTagG5() {
        return this.TagG5;
    }

    public void setTagG5(String TagG5) {
        this.TagG5 = TagG5;
    }

    public String getCourses() {
        return this.Courses;
    }

    public void setCourses(String Courses) {
        this.Courses = Courses;
    }

    public String getDeleteFlag() {
        return this.DeleteFlag;
    }

    public void setDeleteFlag(String DeleteFlag) {
        this.DeleteFlag = DeleteFlag;
    }
}
