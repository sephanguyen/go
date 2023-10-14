package com.manabie.concurrency;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class CustomerService {

    private static final Logger LOG = LoggerFactory.getLogger(CustomerService.class);

    public void updateCustomer(CustomerCSV update) throws Exception {
        // simulate updating using some CPU processing
        Thread.sleep(100);

        LOG.info("Customer " + update.getId() + " updated");
    }
}
