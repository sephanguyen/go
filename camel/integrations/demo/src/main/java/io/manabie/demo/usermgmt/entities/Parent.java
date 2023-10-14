package io.manabie.demo.usermgmt.entities;

public class Parent {
    private String ParentNumber;
    private String ParentName;
    private String ParentRawPassword;
    private String ParentEmail;

    public String getParentNumber() {
        return this.ParentNumber;
    }

    public void setParentNumber(String ParentNumber) {
        this.ParentNumber = ParentNumber;
    }

    public String getParentName() {
        return this.ParentName;
    }

    public void setParentName(String ParentName) {
        this.ParentName = ParentName;
    }

    public String getParentRawPassword() {
        return this.ParentRawPassword;
    }

    public void setParentRawPassword(String ParentRawPassword) {
        this.ParentRawPassword = ParentRawPassword;
    }

    public String getParentEmail() {
        return this.ParentEmail;
    }

    public void setParentEmail(String ParentEmail) {
        this.ParentEmail = ParentEmail;
    }
}
