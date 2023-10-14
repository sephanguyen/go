package com.manabie.transformation.utils;

public class CustomerParser {
    public Customer convert(int body) throws Exception {
        Customer myClassType = new Customer();
        myClassType.setData(String.valueOf(body));
        myClassType.setId("11");
        return myClassType;
    }

    public CustomerParser() {
    }
}
