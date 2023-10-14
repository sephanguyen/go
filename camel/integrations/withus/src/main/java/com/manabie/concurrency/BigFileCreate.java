package com.manabie.concurrency;

import java.io.File;
import java.io.FileOutputStream;

public class BigFileCreate {
    private void createBigFile(File file, int lines) throws Exception {
        FileOutputStream fos = new FileOutputStream(file);
        for (int i = 0; i < lines; i++) {
            String line = "123," + i + ",bumper,4\n";
            fos.write(line.getBytes());
        }
        fos.close();
    }

    public void createBigFile() throws Exception {
        new File("target/").mkdirs();
        File file = new File("target/bigfile.csv");
        if (file.exists()) {
            file.delete();
        }

        int lines = 100;

        System.out.println("Creating target/bigfile.csv with " + lines + " lines");
        createBigFile(file, lines);
        System.out.println("Creating target/bigfile.csv done");
    }
}
