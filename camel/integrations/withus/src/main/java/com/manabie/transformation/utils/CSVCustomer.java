package com.manabie.transformation.utils;

import org.apache.camel.dataformat.bindy.annotation.CsvRecord;
import org.apache.camel.dataformat.bindy.annotation.DataField;

// https://camel.apache.org/components/3.20.x/dataformats/bindy-dataformat.html#_1_csvrecord
@CsvRecord(separator = "\t", crlf = "UNIX", generateHeaderColumns = true, skipFirstLine = true)
// @CsvRecord(separator = "\t", crlf = "UNIX", generateHeaderColumns = true)
public class CSVCustomer {
    @DataField(pos = 1, columnName = "colname1")
    public String customerNumber;
    @DataField(pos = 2, columnName = "生徒メール")
    public String studentNumber;
    @DataField(pos = 3, columnName = "col2")
    public String name;
    @DataField(pos = 4)
    public String password;
    @DataField(pos = 5)
    public String col1;
    @DataField(pos = 6)
    public String col2;
    @DataField(pos = 7)
    public String col3;
    @DataField(pos = 8)
    public String col4;
    @DataField(pos = 9)
    public String col5;
    @DataField(pos = 10)
    public String col6;
    @DataField(pos = 11)
    public String col7;
    @DataField(pos = 12)
    public String col8;
    @DataField(pos = 13)
    public String col9;
    @DataField(pos = 14)
    public String col10;
    @DataField(pos = 15)
    public String col11;
    @DataField(pos = 16)
    public String col12;

    public String getCustomerNumber() {
        return this.customerNumber;
    }

    public void setCustomerNumber(String customerNumber) {
        this.customerNumber = customerNumber;
    }

    public String getStudentNumber() {
        return this.studentNumber;
    }

    public void setStudentNumber(String studentNumber) {
        this.studentNumber = studentNumber;
    }

    public String getName() {
        return this.name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public String getPassword() {
        return this.password;
    }

    public void setPassword(String password) {
        this.password = password;
    }

    public String getCol1() {
        return this.col1;
    }

    public void setCol1(String col1) {
        this.col1 = col1;
    }

    public String getCol2() {
        return this.col2;
    }

    public void setCol2(String col2) {
        this.col2 = col2;
    }

    public String getCol3() {
        return this.col3;
    }

    public void setCol3(String col3) {
        this.col3 = col3;
    }

    public String getCol4() {
        return this.col4;
    }

    public void setCol4(String col4) {
        this.col4 = col4;
    }

    public String getCol5() {
        return this.col5;
    }

    public void setCol5(String col5) {
        this.col5 = col5;
    }

    public String getCol6() {
        return this.col6;
    }

    public void setCol6(String col6) {
        this.col6 = col6;
    }

    public String getCol7() {
        return this.col7;
    }

    public void setCol7(String col7) {
        this.col7 = col7;
    }

    public String getCol8() {
        return this.col8;
    }

    public void setCol8(String col8) {
        this.col8 = col8;
    }

    public String getCol9() {
        return this.col9;
    }

    public void setCol9(String col9) {
        this.col9 = col9;
    }

    public String getCol10() {
        return this.col10;
    }

    public void setCol10(String col10) {
        this.col10 = col10;
    }

    public String getCol11() {
        return this.col11;
    }

    public void setCol11(String col11) {
        this.col11 = col11;
    }

    public String getCol12() {
        return this.col12;
    }

    public void setCol12(String col12) {
        this.col12 = col12;
    }
}
