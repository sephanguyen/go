package com.manabie.transformation.utils;

import org.apache.camel.Converter;
import org.apache.camel.Exchange;
import org.apache.camel.TypeConverter;

@Converter(generateLoader = true)
public final class CustomerConverter {

    @Converter
    public static String toCustomer(int data, Exchange exchange) {
        TypeConverter type = exchange.getContext().getTypeConverter();
        String s = type.convertTo(String.class, data);
        return s;
    }
}