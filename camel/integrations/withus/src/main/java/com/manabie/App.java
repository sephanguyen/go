package com.manabie;

import org.apache.camel.main.Main;

import com.manabie.concurrency.BigFileCreate;
import com.manabie.concurrency.SimpleConcurrency;
import com.manabie.exception.ErrorHandler;
import com.manabie.transformation.CSVMapper;
import com.manabie.transformation.ContentEnricher;
import com.manabie.transformation.Template;
import com.manabie.transformation.TransformByProcessor;
import com.manabie.transformation.TransformMethod;
import com.manabie.transformation.TypeConverter;

/**
 * Hello world!
 *
 */
public class App {
    private App() {
    }

    public static void main(String[] args) throws Exception {
        System.out.println("Hello World!");
        // use Camels Main class
        Main main = new Main();
        BigFileCreate bigFileCreate = new BigFileCreate();
        bigFileCreate.createBigFile();
        // add Java route classes
        main.configure().addRoutesBuilder(TypeConverter.class);
        // main.configure().addRoutesBuilder(UnleashExampleDynamicRoute.class);
        // main.configure().addRoutesBuilder(Withus.class);
        // main.configure().addRoutesBuilder(ContentEnricher.class);
        // now keep the application running until the JVM is terminated (ctrl + c or
        // sigterm)
        main.run(args);
    }
}
