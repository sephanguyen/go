package io.manabie.demo.usermgmt.entities;

import java.util.ArrayList;

public abstract class DomainUser {

    private String UserId;
    private String GradeId;
    private ArrayList<String> LocationIds = new ArrayList<>();


    public String getUserId() {
        return UserId;
    }

    public void setUserId(String userId) {
        UserId = userId;
    }

    public String getGradeId() {
        return GradeId;
    }

    public void setGradeId(String gradeId) {
        GradeId = gradeId;
    }

    public void addLocationId(String locationId) {
        LocationIds.add(locationId);
    }

    public ArrayList<String> getLocationIds() {
        if (LocationIds == null) {
            return new ArrayList<>();
        }
        return LocationIds;
    }

    public void setLocationIds(ArrayList<String> locationIds) {
        LocationIds = locationIds;
    }
}
