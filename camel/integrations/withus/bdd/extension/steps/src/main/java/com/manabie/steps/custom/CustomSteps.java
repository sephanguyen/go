package com.manabie.steps.custom;

import com.consol.citrus.TestCaseRunner;
import com.consol.citrus.annotations.CitrusResource;
import com.consol.citrus.exceptions.CitrusRuntimeException;
import com.consol.citrus.util.FileUtils;

import io.cucumber.java.en.Given;

import static com.consol.citrus.actions.EchoAction.Builder.echo;

import java.io.FileOutputStream;
import java.io.IOException;

public class CustomSteps {

    @CitrusResource
    private TestCaseRunner runner;

    @Given("^MANABIE can be extended!$")
    public void abcd() {
        runner.run(echo("MANABIE can be extended!"));
    }

    @Given("^load Camel K resource file ([^\\s]+)\\.([a-z0-9-]+)$")
    public void loadIntegrationFromFile(String name, String language) {
        try {
            String fileName = name + "." + language;
            String data = FileUtils.readToString(FileUtils.getFileResource(fileName));
            FileOutputStream outputStream = new FileOutputStream(fileName);
            byte[] strToBytes = data.getBytes();
            outputStream.write(strToBytes);
            outputStream.close();
        } catch (IOException e) {
            throw new CitrusRuntimeException(
                    String.format("Failed to load Camel K integration from resource %s", name + "." + language), e);
        }
    }

}