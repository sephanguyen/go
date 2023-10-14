#!/bin/bash

# This is generated from skaffold.manaverse.yaml using draft-context, but with hasura disabled and
# some objects removed so that each kind only has one of its object available.
# The most important thing is to ensure our tool and parse a valid templates with various kinds of objects.
cat <<EOF
---
# Source: manabie-all-in-one/templates/limitrange.yaml
---
apiVersion: v1
kind: LimitRange
metadata:
  name: cpu-limit-range
spec:
  limits:
  - defaultRequest:
      cpu: 10m
    type: Container
---
# Source: manabie-all-in-one/charts/draft/templates/app.yaml
apiVersion: policy/v1beta1
kind: PodDisruptionBudget
metadata:
  name: draft
spec:
  maxUnavailable: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: draft
      app.kubernetes.io/instance: manabie-all-in-one
---
# Source: manabie-all-in-one/charts/draft/templates/app.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: local-draft
  labels:
    helm.sh/chart: draft-0.1.0
    app.kubernetes.io/name: draft
    app.kubernetes.io/instance: manabie-all-in-one
    app.kubernetes.io/version: "1.16.0"
    app.kubernetes.io/managed-by: Helm
  annotations:
    
    iam.gke.io/gcp-service-account: local-draft@.iam.gserviceaccount.com
---
# Source: manabie-all-in-one/charts/draft/templates/app.yaml
apiVersion: v1
kind: Secret
metadata:
  name: draft
type: Opaque
data:
  service_credential.json: |-
    ewogICJ0eXBlIjogInNlcnZpY2VfYWNjb3VudCIsCiAgInByb2plY3RfaWQiOiAiZGV2LW1hbmFi
    aWUtb25saW5lIiwKICAicHJpdmF0ZV9rZXlfaWQiOiAiYjk0MjkxOThkNzk5OTdmZGE2ZDA5M2Jj
    ZWJlODFhN2Q3YWI2OTkzOCIsCiAgInByaXZhdGVfa2V5IjogIi0tLS0tQkVHSU4gUFJJVkFURSBL
    RVktLS0tLVxuTUlJRXZBSUJBREFOQmdrcWhraUc5dzBCQVFFRkFBU0NCS1l3Z2dTaUFnRUFBb0lC
    QVFDNkVoTm5Mc0NxdUlhTFxudlRsQk9VK080cTRyOFpIekx6WEQ3SXF0UHhxNXVSbTdJNUgxZU94
    TU9DUXp6Z1dkVFZJaHZ5Y2h1eXZMQytWcVxudk92MytSUjRaVWVqbUxMSnVGckREMnBXUHIyTnZt
    MVgxcjU3VFRsZElyMWMzUWNmNWlwU3pRMm42Z1g5TnQ4RFxuSEVrZzNzdlJJUjRmQytOb3JGUVlJ
    UkZVOTUrUjU0VWk1czVHaEhZWVpUVDE5b2lKcm0rUUVSNTdJZlBaWFlRclxuOVFkQmpCQTZCMGpy
    OXdseVRMQWplRkw0U1RFYm94aWxmd01iU2o4MHNRa2FyL0tiUjJoQnlpbmxmL0xUZ0N3elxuc2tr
    QmsrbGxiMmM3blpLaWgyeDV6cGVoQWdla0d6ZW9BQjZRbExxTTJtUDZVUlNPK0tkUVA3cENDaDBS
    UXhvbFxuR01ZNHgxaVRBZ01CQUFFQ2dnRUFCNFZkaFdrdFhua3c3d3NKK21udm5rM3BUbHRvVTlV
    UHJraXNYazVUclRnZlxuSXlKUDd3VWhQLzl3N3lzZnJQa0lIZGNWSk5iazhVTWMxZENuRlJIYlV2
    WjlDODdMUXo0UlpSc0ZhRkVHNW1qUlxuRUtEY2VDMXA2U3JUVHFLY2ZCeVlqMW84ZUJJTWhleW0z
    UUJTc0dKeENKWDNHcmduUy83VE0xcDYwZDFrZE1nK1xuRHp3L2o0YittTXVUenI3cGtnN2hicGRU
    WGtyb0swemY5dDVTSFpKaS9XS1A5b0R3Zkh4Nks2MnRmQ05sdTFRaVxuYXNxR09HNksybHNlL1Jp
    VEE0enFFNUJnZmFIOG95eGFqMXlEOVNNN2JUYVdrVU1mZExLaU1HVGZpUEFXQXpNYVxuM3BUZnpy
    VDlpYUNXWWg1eWI1cDR4cUZzWEM0YWJvU1JHdyt1dFM1RStRS0JnUUQrSFdTenNNcGNOMGp3UFJO
    MlxuVUlHMGpHell2NmhIWmxBRCtwQWFjaW50akxKRFo1ZU1mL1dGWWdGMWpSQldNVEROOUp3bFFj
    dU9yTHlLSjV2dFxuaDFQN2MxQ0prbzVWSFpuWXc5NHZEbm9zWldLTW1yeTZMYzRKbjh1RWkwU0lQ
    c2NzL0s4UVF2K1FibDV5bmFHYVxuVktUY2dnYkdMb1h3ektNZzhxeUpSZzh3cHdLQmdRQzdjM1Iv
    VGtHVUprbTZUVnA4NnNHQ1ZtT2Z2Y1NmTUk5Y1xuRXdXSzdJVXBYZ2lrWXlOZytmcUZyZHhyQTd0
    VS83aGhhSnJlVmtmRk9NOG8wblMvQUcrOWpUTDZHaEs4MEMvTVxuWEg1QkJwcUJWSU1wc2RIOWZJ
    c2JQUS92UCtsNVI2WDhaWkV2K1cyMk1BK0M3ajMzMnZQS3N2TmZsQm1GQlRxaVxuV25QTVlCNUtO
    UUtCZ0JJNVVXdUJsa0dleFdCVlFQd1BNZjRjeEFHWFhSNGh2RU5NeU9EY3B4MGVKZnFuaHpyUVxu
    UW05YVkvaG1NWEc4L1Y4SDE5cmtLUkVHV2s4ZUlCU2N5KzBRakFvUnRKdHVFQVozcFl1Q1lraWt6
    TGlBc0dBNVxud0xqMytNUjhxR0dNL3dPKzYxOGpMdWpRd1gwK3lNUWtwZDRhaFJuWlpFbXNvMVpO
    a1FvWE9DZXBBb0dBUnQ3dFxuNnJ2aG0ydW1jR09TbEt3RklZd2IrbWM3RVp6QWR1VlNNU1lmYW5a
    OCtmbnBoRjYrMHcvYXlETU8vcUg0U2d2TVxua2NjNU4xMjFKUS84eDhJWWZTZ0hYL3UvbmRkd1d1
    bVZhbXhldWdzRDFCM0E4UC9IY0RMejlWYktwT25yM2JOZ1xuNHl5QXlHTC9XbGRNNG9yTHBaVm00
    bm9SOC9MNEtpM2NuaWF4RFFrQ2dZQThSSzZOanpPRmxRNDVPYy9SOXcwTFxualhmbS9WZFh4Tmdp
    ajg3V1RYS3lsUzREbWsvYmhzVmlDQ2syWmtFN1J1bzdpcTl1b3JCbnZDdFVhUEt5aFNTaVxuNkdJ
    Zm1ZbzRHUzIvRSt1c2FLZlZUcnUvNVI2bFEzd3drSFZwR055bVZldTdYRXF5NXVJMmhTSUtqMTlW
    NzloSVxuNjNhOGN3bklIdTVKbFJPTkNWWXByZz09XG4tLS0tLUVORCBQUklWQVRFIEtFWS0tLS0t
    XG4iLAogICJjbGllbnRfZW1haWwiOiAiYm9vdHN0cmFwQGRldi1tYW5hYmllLW9ubGluZS5pYW0u
    Z3NlcnZpY2VhY2NvdW50LmNvbSIsCiAgImNsaWVudF9pZCI6ICIxMDQxMDMxNzM4OTY0MDgwNzk1
    MzEiLAogICJhdXRoX3VyaSI6ICJodHRwczovL2FjY291bnRzLmdvb2dsZS5jb20vby9vYXV0aDIv
    YXV0aCIsCiAgInRva2VuX3VyaSI6ICJodHRwczovL29hdXRoMi5nb29nbGVhcGlzLmNvbS90b2tl
    biIsCiAgImF1dGhfcHJvdmlkZXJfeDUwOV9jZXJ0X3VybCI6ICJodHRwczovL3d3dy5nb29nbGVh
    cGlzLmNvbS9vYXV0aDIvdjEvY2VydHMiLAogICJjbGllbnRfeDUwOV9jZXJ0X3VybCI6ICJodHRw
    czovL3d3dy5nb29nbGVhcGlzLmNvbS9yb2JvdC92MS9tZXRhZGF0YS94NTA5L2Jvb3RzdHJhcCU0
    MGRldi1tYW5hYmllLW9ubGluZS5pYW0uZ3NlcnZpY2VhY2NvdW50LmNvbSIKfQ==
  draft.secrets.encrypted.yaml: |-
    cG9zdGdyZXM6CiAgICBjb25uZWN0aW9uOiBFTkNbQUVTMjU2X0dDTSxkYXRhOks0ZFJYcmltWDdsQVNUS0FPQldEQTN3anhlMzlOK0ltWHRGeWp5Z2pJb0szTkxoamFZN0UwcDdBYzVHVDN2TERpYTRtZG1EQXRpTkZWL2JUM2FXTDZrd0hjS2w3alhYTzRmVkRQNU9udTFDQVBKTzROUWMyMUJGYmlIRkdJZz09LGl2OlRIMnUxMmk3VnFSeU9XZHluSTBhSmwrUlhKRnpMNjQrcGNLRisraWpLV1E9LHRhZzpSTnJCZE9rOU9LZ05qTkptZkUzR1dRPT0sdHlwZTpzdHJdCnRvbV9wb3N0Z3JlczoKICAgIGNvbm5lY3Rpb246IEVOQ1tBRVMyNTZfR0NNLGRhdGE6YnNiYjdleE5YbzRGK1dSSURJREVaY3k1QmtKbEFuSDJ4MHNQd3R1ZnBMMG5uWjF0TDBsQ0xPdVNWU3p3MVdGamQ4QWdlWnVIMGdWWFkraEROSmhQN2xoL3g1ckpOdTZnS2FQYUxid1RNQmV4MU11R29YNG9ONEhNRXdCS2VtOD0saXY6UVFvVGs0T1JkWmJUQUl4bFhXNmFqMEFJa3E0QW1BcDFxSTBxNGhvdzI4UT0sdGFnOmZQc3VJNUJQVGpGTGJ1dlJZeW43VXc9PSx0eXBlOnN0cl0KYm9iX3Bvc3RncmVzOgogICAgY29ubmVjdGlvbjogRU5DW0FFUzI1Nl9HQ00sZGF0YTp5UnZTRDBSenM4QnQ0QytwNXA1YzYwWE9TMU94czZNbDJlZmg3Z3JzQmc0cEliM0pvMU94SUxvdmw5ZWJFZmR5R2VyRkxvRUR1aGdSOFc4WkhjUU9WVzZhK1pRNC9ZMzZWam0wS1BDVG1EZ2ZwZ0VGQTJCR1hyNllzVU0xODlNPSxpdjp2cHNGKzQ2dDU4TGlzZURydmRDRDdXeWRzcEUvOWs1dkRwMlpHMGtOaDZ3PSx0YWc6bEdTUG5QMVRXMlFUR3RpcFROM3VqQT09LHR5cGU6c3RyXQpldXJla2FfcG9zdGdyZXM6CiAgICBjb25uZWN0aW9uOiBFTkNbQUVTMjU2X0dDTSxkYXRhOkJzdXVxUVoyRllFemppQUpjamVzckFUWndYeFlZUS9KT0pIQU16WmlqY1h2TnNKK3ROT2ZlMk9nSno5cTBPQ1h5U25oVVhxd0FLbGsvcUJvcHVCcSt4cVJ0b3A3VUJEYlg5empjL1FoZDRwNzkrc2xOS0JGTk5CTElOVytqN2lVUWwwPSxpdjpxM3U3Y29ZOEh0clZjUGRucTVxbTQyeVJwZjIySlFEdTlnZTgrYVNyMW1VPSx0YWc6bFBKdTJQUHZFTW1CY293YXRucFkrdz09LHR5cGU6c3RyXQpnaXRodWJfcHJpdmF0ZV9zZWNyZXQ6IEVOQ1tBRVMyNTZfR0NNLGRhdGE6SUdSRGxnWkJBSjJNeXM4VnA5dU1JZWxvTjBSSzhRRExXbUhheXN5eGpVSnEzSEJVZjdDRDV2ZGZrNUpERFYrYU5qQUwzYWllYTlmYk9UYi9YM2lKZVFJTE1FVFRIWnBKK3RudzNaNnhpN1MwaDluSDVUdko4N0xzeVNuTTNxV0tPdERTTGdFdHp5Q1ppbW5LYnBLaE43TVZVaXA4Qzl4bGZ5bkZrbklJaVZtc2FpVEZTb0t3YnMyYndGRFhGcngzYXdib0huaHJZUXdUNHVkQ29HOE4yWlpFQmg4cVZEM3V5NDZJZFAxb0d5VmpMZmNqLytLcTNoclBLc0NmY3FBc0VUM0tqV3doMDlpQUtqY0pRM1NVNXhhQkpoMVlTTzRNMWNablp0dzdLVEUzVHliV0Q3QXRQMDhjQjhjSkszdjU0ZkpJUklXc2tYS1A4OGdPM2JXWWtDR0hWTE1YUndxV1M1dmQ3RUNBc1hpS0Rzclo5K2FTWE9PTEZKRVYyWU1EZEEyQm1nSnZIZitsMDlLRkN6OFJKRWdEYjIvVjdUdTA2b0pPZ1JKUUhBMGNocERDbUh0akg1MnprUGJBdlNVQ1ljRGgvYzBBVlI4cGRyTW8xUW9FSU50SkIwaFdUQWM3NXB6eTBTa2M0NXdOb2lrZll1ZUtVMXdVczlGMUhTZ1hKK09JbGJyOWRTR25HUzFEdVNuSG9kMkZYbUFoMS8yV3BEZjBIcXl0Ym1teWVicU5YVEJ1R01iNEFiSGZreWh1b2NoR3IzTkhyODhlZ1lUU2FuUU5KR3RUL2Yya2tyZXB4RVJBM1Bja3NFT1dSR3RxeC9rWVMrQ2J6ZVNWU3lobm4vRVoxNmc1angwMDFPYTBQYm1WbjlvN2hWY3Jteld5U3ppLzY3RW45aUxxdndLbDNOL0ZhTjFNaGVjZ3ZpcHNSeUYrR0s0dnZHZUNnaVlDUDF5STBkWnpBNTZudCtGeFhHK0VGM3BjVUkwS3dxZWEyVVlndmN2Sm1XRzZGamVDd2ZrSC9wR05Ra0ZhT1V5Sk5qODd0L0kvZCsrbkxpdjhYMWlnZlE1QnBrU2E1Yks1bDhpV2J0QkRFQVNvOTAzVzJsNXdjL2ZXMWpHdlZoWDlmNU9zR1pub0RlaERFTEJzd2JXRDQ3ZUUvMFhpUDZuRm9DVWhQRUtBYytLbW9jZlhMc1FFRmI4N05ZNGlLYmkyTDZyV0ZFNlY3VUNlQlA0QWpDU1Vrb0xKMEFDUlZNalIwWDh4bWV4c3QwZlY2RFk0bTl5UnorbzAvNSs4OW9jSlVzRmlTYTlKMUdBNUVwenZqUWVEZ0VvVWhrK08xV1NJNHZNMGJZbjZlT2xPYUNkWkxKZjMyb3Bod0RLVEJnUmVPZ3k0Yi85RkF5ODFld05CQXozWW5SMW1vai9SUFRpVzRCdGZrUnBsTFRyWm05RUN1aXdRdmFodEp0UjlsQTVrQU9jQVV0TEFIdWZIOVlxbHJRNVBTS28zdGlnTUFQSjl4TUFUVW8rL01qakVCcTdyZ1pUalZHSENic2c4Q0FIZVRZZVdQN3dtK2g4K3o4RlJRdzljS3lLSEVaL0Rxd3ZWYXNKSkhRVm53VFhpdlljU1Q2bDYvSEJ3aDFDRjkwdm9EeCtmU1oyeDZ1ZmZQUkFENTdOMEF6NWhtUkpUN2xGdWtzeFMwVTRUNlhjazNVb2tQb1djWjY3L2luRmtPZkpkQ0VDV2lYaFR4VzVkcXduTy9wM0hiT0pJKy9kcmt6OUdlK243bVlJZzhXeUVjbXVqTVBGUFhHMVJ5SHJqck1pMkl2dUl2Y09mbjhnaHQvamlqNVBFYnlFb25aaUhyQzNpSDREb2l3ZU9OQnM4YVM5MVVBalFzcUl5N0tKWWlBck94SjJYTDdLZTdGVC9ieFc2QVdBdko0UmNVdHFuNHlhTFUzUzE5R056OGhNOEJlbGVPWVlHU3JKZkZJeE5WWi9kN2ZJZHNjZGZUYjFtb256UjJwb2h2YnU4QmRuR2t6Q2wxZXdiZnJhSWxrN2RWK0FBYmdqRm91OFp1VFh3OHA3cmhDa3l2czBWMzFaOUFkbUxxZXZUeWI3N05sYm1nRTZ4Q2twWVdHRlhBb09GdmZLSUdCaUZROWwzbzNkc2laM1V3L1pub21LRmFyTys1RGdISnh3SlJaZ0N4clA4YkIvV052YzJSREtKUGRPVUMzQmxyN2dkR1EycGFPKytIMzVvcjJWOTVKTkxRNU8zSVVnR2RKeElPRjhjeVdQZ1plakwySEQrc3pnR0VuWFhQc1F2R2Y4aTlBRXpEV3JrZGd0cVROSThZalN1WCtEeTBwTlAwMG1rSUhBbnEzZzV1R2NiSDJGTzdzQ3RycTJhcXVMbGtHWStDWFlvUUJQZ1FuU3ozb2tOMUlXNnJjR0hyY3RGL1VBRlg1U3RpRTNpOUM1WUlmZFkwRXJDN09FQjhaK1hvbGxDRkJ3K2R4NjBuOFlzNzV1SVY1eWdZbFliQ1hsSWMvSzJwQTMwS2t6dTdVc2M3L2QvOUhtYmxUR0FRdTFaMWNQMThSVXlyb21lZVc4M1h1QU0ycHJJOEZ0VlBPMFo4Q0ppMFNmZW1VUG8zbjZOWC9OalJTQnNXRm1wVXJqM0JZc0MyYXlhY3Izc0dqNG9KNHdjV3RyNW9lV1FkTjU1YWJBOXlTUHV1RWwzcDBVSWN3VktNcm1zUVo3bVFJc0VXN0NPWXVyNjc4TEM5bTFWdDJUSjBmeGNIdCtnajBGWDljcWtGOVVLZmxHWE1PTGpScVpwSEoxNEQ0aEJjSzlVZDZuS2gyUGNhM3FWazlKNjl1U2czUVl5ZTlQL2g2R1NRblp5L0Z5N2FBbHR3VlNWWXZzRFZPa2RBdTF5bEJoREcrdTEwUDlScDVzQk5VQThra1J6TnNWT1Ivb051WGhvVXhONUtjUDFXT1BKK2hpTjNVRWlDOGhzLGl2OlB2R3ZGMkY3V2JQWGVXUHd5N2dlT003UHRiMGs0TTNnUktmb3dVTXRwcGc9LHRhZzpkSXJIS2h2ZzRKVDlWQUUwVjBWdXR3PT0sdHlwZTpzdHJdCmdpdGh1Yl93ZWJob29rX3NlY3JldDogRU5DW0FFUzI1Nl9HQ00sZGF0YTpsL0Jlc2tiZk1VVTRVNjNuL1E9PSxpdjprK2VBZ05vaElJNDBRSWtVdnJ3RERvVUxXU09neGRhZ2VoYTl3bDZNQTBRPSx0YWc6M3FPTDdQU1pGL3FOd0NQQzdtekhzdz09LHR5cGU6c3RyXQpzb3BzOgogICAga21zOiBbXQogICAgZ2NwX2ttczoKICAgICAgICAtIHJlc291cmNlX2lkOiBwcm9qZWN0cy9kZXYtbWFuYWJpZS1vbmxpbmUvbG9jYXRpb25zL2dsb2JhbC9rZXlSaW5ncy9kZXBsb3ltZW50cy9jcnlwdG9LZXlzL2dpdGh1Yi1hY3Rpb25zCiAgICAgICAgICBjcmVhdGVkX2F0OiAiMjAyMi0xMS0wOVQwNjowOTowNFoiCiAgICAgICAgICBlbmM6IENpUUEvMmxIalhielcrZWgwZDJJRU1TVG9JNXRjRHh5VFR1cFFNZ0paOFdtdnVPTkdSTVNTUUQxQngvSUlUYm5hN2VoWU5KT3BCU2RuZXNsNWRnQ0k4OThpblBXK2ZHV29FVGh4RTJ1Y2VXSVdaeG9nS1RCSU8xSGJTWjgxamg0UzRjRnJpQmc0cEdHTnB4NDVxUkhBWHc9CiAgICBhenVyZV9rdjogW10KICAgIGhjX3ZhdWx0OiBbXQogICAgYWdlOiBbXQogICAgbGFzdG1vZGlmaWVkOiAiMjAyMi0xMS0wOVQwNjowOTowNVoiCiAgICBtYWM6IEVOQ1tBRVMyNTZfR0NNLGRhdGE6cFoxczkvTVBoMHNOM2pZTVZTL3VwdnMzT0Y1QnhHbWZEenVhK3pqQWtvVkdRYndaVkxTWEZ6VGE1b2hLUDZjczlmTVBncGZuOTRlaGxXdHFFYjh3c0c3aTd3N3lNOVZ5dlJHcVZTVSt3ZWJscUhMbkpZbHhENHNEK0ZjRVhsUXpWOWd4OUVjZVVXRmVGdzM3S0VicDU1SGliUHlBeVBxWGtxVVdQSHZ6cXp3PSxpdjpsODNKbDdSNnVJYnZSa1RhM3BLdGxLb2lVdktjQXJqUll4bldxM0gwTUhFPSx0YWc6Q05yeXBoWGh6bS9lejg3YzZtV2Nxdz09LHR5cGU6c3RyXQogICAgcGdwOiBbXQogICAgdW5lbmNyeXB0ZWRfc3VmZml4OiBfdW5lbmNyeXB0ZWQKICAgIHZlcnNpb246IDMuNy4zCg==
  draft_migrate.secrets.encrypted.yaml: |-
    cG9zdGdyZXM6CiAgICBjb25uZWN0aW9uOiBFTkNbQUVTMjU2X0dDTSxkYXRhOmowZUFhRGhrclUxOGFEYWh1cnVOMlp0TVdaSlVTeWNTU1ljM3dYdnNUMU4xWGtKMHpWaXRmeWtsSjEzVTVtQUxqb1RaT0MyL3dWWkFueS9FM0JQdnZhRFJDMGVjd21OMjcxVmZSRGJVOWd3QlE5WEJhbS9DdHkvZDZ1R1RSK2NYSVE9PSxpdjpaYTVwb2kvMU1EQ05EWWdWcEZOTVVFTjBFMDZ3dEVhTXZqY21aZnZOMzgwPSx0YWc6RnZkcDFkLzJmOGFuYTgxUDBJazFjdz09LHR5cGU6c3RyXQpzb3BzOgogICAga21zOiBbXQogICAgZ2NwX2ttczoKICAgICAgICAtIHJlc291cmNlX2lkOiBwcm9qZWN0cy9kZXYtbWFuYWJpZS1vbmxpbmUvbG9jYXRpb25zL2dsb2JhbC9rZXlSaW5ncy9kZXBsb3ltZW50cy9jcnlwdG9LZXlzL2dpdGh1Yi1hY3Rpb25zCiAgICAgICAgICBjcmVhdGVkX2F0OiAiMjAyMi0wOC0wOFQwODo1ODo1MloiCiAgICAgICAgICBlbmM6IENpUUEvMmxIamFmanlpM2M0b2xGQWN1QVFEVEVHMEdQSVBmWTVsZ2hpM3ZzaTVNclBwY1NTUUJDanMzdEZKY1ZYUTFpR2dZckQ2TmF0NVNhMGpSdEFzeDdyWmYzTlBLV2dwWE9QcXZ1cGNFSjBPOFJwTm1XekFXOUI0MWNFTkppTVZJUFhoalowOFg4SlRoN1JoUlRQaEU9CiAgICBhenVyZV9rdjogW10KICAgIGhjX3ZhdWx0OiBbXQogICAgYWdlOiBbXQogICAgbGFzdG1vZGlmaWVkOiAiMjAyMi0wOC0wOFQwODo1ODo1MloiCiAgICBtYWM6IEVOQ1tBRVMyNTZfR0NNLGRhdGE6VE5qdncxL1RxajNtVzFBaFRNSDh3VUtDMDRyck9YT0lHVEVjL1l6YXR4Ymdramt1NkpyYWcvdHFTT1kvUlJzUVNCSjZlQ3RDZjRYS3FKY1dSNWlSelk4Qk92aXR3VWE0c0JWRSt1YlViTG1aMk9iTzFOYTFlVCtJdDJyZ1hqS1RpUDVqN2RBOFdVVDlxVFFHeEJ0TS9Hd0V5RWlqSXFOdmoxT2NVZGt5aTB3PSxpdjpRR3RCNFBwUHdsL2xxMHh2UlUrdTNBTHJSSTViRDUwZHNHVG1XSHljUC80PSx0YWc6UUR5eWNWQzlzdktaWjdSMXMySnlMZz09LHR5cGU6c3RyXQogICAgcGdwOiBbXQogICAgdW5lbmNyeXB0ZWRfc3VmZml4OiBfdW5lbmNyeXB0ZWQKICAgIHZlcnNpb246IDMuNy4zCg==
---
# Source: manabie-all-in-one/charts/draft/templates/app.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: draft
  labels:
    helm.sh/chart: draft-0.1.0
    app.kubernetes.io/name: draft
    app.kubernetes.io/instance: manabie-all-in-one
    app.kubernetes.io/version: "1.16.0"
    app.kubernetes.io/managed-by: Helm
data:
  draft.common.config.yaml: |
    common:
      name: draft
      sa_email: local-draft@.iam.gserviceaccount.com
      log:
        app_level: warn
        log_payload: true
      stats_enabled: false
      image_tag: locally
      listener:
        grpc: :6050
        http: :6080
    postgres:
      connection: postgres://draft@.iam@127.0.0.1:5432/draft?sslmode=disable
      max_conns: 2
      log_level: debug
      retry_count: 10
      retry_interval: 5s
    bob_postgres:
      connection: postgres://draft@.iam@127.0.0.1:5432/bob?sslmode=disable
      max_conns: 8
      log_level: debug
      retry_count: 10
      retry_interval: 5s
    tom_postgres:
      connection: postgres://draft@.iam@127.0.0.1:5432/tom?sslmode=disable
      max_conns: 8
      log_level: debug
      retry_count: 10
      retry_interval: 5s
    eureka_postgres:
      connection: postgres://draft@.iam@127.0.0.1:5433/eureka?sslmode=disable
      max_conns: 8
      log_level: debug
      retry_count: 10
      retry_interval: 5s
    clean_data:
      conversations:
        ignore_fks: ["last_message_id"]
      users:
        extra_cond: |
          and email not in (
            'trungkim.tran+virtual_teacher01@manabie.com','trungkim.tran+virtual_teacher02@manabie.com','trungkim.tran+virtual_learner01@manabie.com','trungkim.tran+virtual_learner02@manabie.com',
            'schedule_job+notification@manabie.com','product.test+jprep.staging@manabie.com','phuc.chau+e2ehcmschooladmin@manabie.com'
          ) and email not like '%thu.vo+e2e%' and email not like '%schedule_job+%'
      media:
        extra_cond: "and media_id not in ('01GBPSXE5XXY0N49080Y5X8N6D','01GBPT7QQCRPW1RVA1W4Q3ZZ5K','01GBPSQR1E5CH7GFH4X65BNA26','01GBPSQR1F9RTPBXN70NT7FYQZ','01GBPSQR1F9RTPBXN70RPNR2WB','01GBPSQR1F9RTPBXN70S1JHP5X','01GCTEG1F4017JCX881SXK9NPN','01GCTEG1F4017JCX881W0V2NX7')"
      lessons:
        extra_cond: "and lesson_id not in ('01GBPSQRDARCHZG07XNDVC8HTK','01GBPSXEGN0HCG3DPB3FASEF99','01GBPT7R14XYQ4NG267HRWKGK5')"
      courses:
        extra_cond: "and course_id not in ('01GBPS6YPZ800YDZPBZKXJZ7W9')"
      locations:
        extra_cond: "and name not like '%E2E%' and name!= 'End-to-end'"
        ignore_fks: ["parent_location_id"]
        self_ref_fks:
          - referencing: parent_location_id
            referenced: location_id 
      study_plans:
        ignore_fks: ["master_study_plan_id"]
        self_ref_fks:
          - referencing: master_study_plan_id
            referenced: study_plan_id
      student_submissions:
        # this is a circular fk, we ignore one (typically ignore the nullable fk)
        ignore_fks: ["student_submission_grade_id"]
      student_submission_grades:
        # if have conflict on this we set null using the query () 
        set_null_on_circular_fk:
          student_submissions: student_submission_grade_id
    
    
    
  draft.config.yaml: |
    common:
      environment: local
      stats_enabled: true
      log:
        app_level: debug
      remote_trace:
        enabled: true
      grpc:
        trace_enabled: true
        handler_timeout: 5s
        client_name: draft
        client_version: 
      google_cloud_project: dev-manabie-online
    issuers:
      - issuer: http://firebase.emulator.svc.cluster.local:40401/fake_aud
        audience: fake_aud
        jwks_endpoint: http://firebase.emulator.svc.cluster.local:40401/jwkset
      - issuer: manabie
        audience: manabie-local
        jwks_endpoint: http://shamir:5680/.well-known/jwks.json
    github:
      app_id: 246581
      installation_id: 30125001
    postgres:
      log_level: debug
---
# Source: manabie-all-in-one/templates/tests/rbac.yaml
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: "local-manabie-tester-cluster-role"
rules:
- apiGroups:
  - ''
  resources:
  - pods
  - secrets
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - networking.istio.io
  resources:
  - virtualservices
  - gateways
  verbs:
  - get
  - list
  - watch
---
# Source: manabie-all-in-one/templates/tests/rbac.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: "local-manabie-tester-cluster-role-binding"
subjects:
- kind: ServiceAccount
  name: "local-manabie-tester"
  namespace: backend
roleRef:
  kind: ClusterRole
  name: "local-manabie-tester-cluster-role"
  apiGroup: rbac.authorization.k8s.io
---
# Source: manabie-all-in-one/charts/draft/templates/app.yaml
apiVersion: v1
kind: Service
metadata:
  name: draft
  labels:
    helm.sh/chart: draft-0.1.0
    app.kubernetes.io/name: draft
    app.kubernetes.io/instance: manabie-all-in-one
    app.kubernetes.io/version: "1.16.0"
    app.kubernetes.io/managed-by: Helm
spec:
  type: ClusterIP
  ports:
    - name: http-port
      protocol: TCP
      targetPort: http
      port: 6080
    - name: grpc-web-port
      protocol: TCP
      targetPort: grpc
      port: 6050
  selector:
    app.kubernetes.io/name: draft
    app.kubernetes.io/instance: manabie-all-in-one
---
# Source: manabie-all-in-one/charts/draft/templates/app.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: draft
  labels:
    helm.sh/chart: draft-0.1.0
    app.kubernetes.io/name: draft
    app.kubernetes.io/instance: manabie-all-in-one
    app.kubernetes.io/version: "1.16.0"
    app.kubernetes.io/managed-by: Helm
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: draft
      app.kubernetes.io/instance: manabie-all-in-one
  template:
    metadata:      
      annotations:
        checksum/draft.common.config.yaml: e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
        checksum/draft.config.yaml: e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
        checksum/service_credential.json.base64: fc76925ae89d1194c7a78786e24356f7f3e5edcbe2028859b2baa9340fafa542
        checksum/draft.secrets.encrypted.yaml: e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
        checksum/draft_migrate.secrets.encrypted.yaml: e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
        cluster-autoscaler.kubernetes.io/safe-to-evict: "true"
        sidecar.istio.io/proxyCPU: "10m"
        sidecar.istio.io/proxyMemory: "50Mi"
      labels:
        app.kubernetes.io/name: draft
        app.kubernetes.io/instance: manabie-all-in-one
    spec:
      serviceAccountName: local-draft
      volumes:
        
        - name: secrets-volume
          secret:
            secretName: draft
            items:
            - key: draft.secrets.encrypted.yaml
              path: draft.secrets.encrypted.yaml
            - key: draft_migrate.secrets.encrypted.yaml
              path: draft_migrate.secrets.encrypted.yaml
        - name: service-credential
          secret:
            secretName: draft
            items:
            - key: service_credential.json
              path: service_credential.json
        - name: config-volume
          configMap:
            name: draft
            items:
            - key: draft.common.config.yaml
              path: draft.common.config.yaml
            - key: draft.config.yaml
              path: draft.config.yaml
      initContainers:
        
        - name: draft-migrate
          image: localhost:5001/asia.gcr.io/student-coach-e1e95/backend:locally
          imagePullPolicy: IfNotPresent
          command:
            - /bin/sh
            - -c
            - |
        
              set -e
              # run migrations
              startTime=$(date +"%T")
              echo "draft-migrate start time: $startTime"
              /server sql_migrate \
                --commonConfigPath=/configs/draft.common.config.yaml \
                --configPath=/configs/draft.config.yaml \
                --migratePath=file:///migrations/draft \
                --secretsPath=/configs/draft_migrate.secrets.encrypted.yaml
              exitcode=$?
              endTime=$(date +"%T")
              echo "draft-migrate end time: $endTime"
        
              duration=$(date -d @$(( $(date -d "$endTime" +%s) - $(date -d "$startTime" +%s) )) -u +'%H:%M:%S')
              echo "Duration: $duration"
              exit $exitcode
          volumeMounts:
          - name: config-volume
            mountPath: /configs/draft.common.config.yaml
            subPath: draft.common.config.yaml
            readOnly: true
          - name: config-volume
            mountPath: /configs/draft.config.yaml
            subPath: draft.config.yaml
            readOnly: true
          - name: secrets-volume
            mountPath: /configs/draft_migrate.secrets.encrypted.yaml
            subPath: draft_migrate.secrets.encrypted.yaml
            readOnly: true
          - name: service-credential
            mountPath: /configs/service_credential.json
            subPath: service_credential.json
            readOnly: true
          env:
          - name: GOOGLE_APPLICATION_CREDENTIALS
            value: "/configs/service_credential.json"
      containers:
        
        - name: draft
          image: localhost:5001/asia.gcr.io/student-coach-e1e95/backend:locally
          imagePullPolicy: IfNotPresent
          command:
            - /bin/sh
            - -c
            - |
              #!/bin/bash
              set -eu
              cat <<EOF > modd.conf
              /server {
                daemon +sigterm: /server \\
                  gserver \\
                  draft \\
                  --commonConfigPath=/configs/draft.common.config.yaml \\
                  --configPath=/configs/draft.config.yaml \\
                  --secretsPath=/configs/draft.secrets.encrypted.yaml
              }
              EOF
              exec modd
          volumeMounts:
          - name: config-volume
            mountPath: /configs/draft.common.config.yaml
            subPath: draft.common.config.yaml
            readOnly: true
          - name: config-volume
            mountPath: /configs/draft.config.yaml
            subPath: draft.config.yaml
            readOnly: true
          - name: secrets-volume
            mountPath: /configs/draft.secrets.encrypted.yaml
            subPath: draft.secrets.encrypted.yaml
            readOnly: true
          - name: service-credential
            mountPath: /configs/service_credential.json
            subPath: service_credential.json
            readOnly: true
          env:
          - name: GOOGLE_APPLICATION_CREDENTIALS
            value: "/configs/service_credential.json"
          ports:
            - name: http
              protocol: TCP
              containerPort: 6080
            - name: grpc
              protocol: TCP
              containerPort: 6050
          readinessProbe:
            exec:
              command:
                - sh
                - -c
                - /bin/grpc_health_probe -addr=localhost:6050 -connect-timeout 250ms -rpc-timeout 250ms
            initialDelaySeconds: 10
            periodSeconds: 5
            timeoutSeconds: 5
            successThreshold: 1
            failureThreshold: 5
          resources:
            requests:
              memory: 128Mi
        - name: draft-scan-rls
          image: localhost:5001/asia.gcr.io/student-coach-e1e95/backend:locally
          imagePullPolicy: IfNotPresent
          command: 
            - /server
          args:
            - rls_check
            - --commonConfigPath=/configs/draft.common.config.yaml
            - --configPath=/configs/draft.config.yaml
            - --secretsPath=/configs/draft.secrets.encrypted.yaml
          volumeMounts:
          - name: config-volume
            mountPath: /configs/draft.common.config.yaml
            subPath: draft.common.config.yaml
            readOnly: true
          - name: config-volume
            mountPath: /configs/draft.config.yaml
            subPath: draft.config.yaml
            readOnly: true
          - name: secrets-volume
            mountPath: /configs/draft.secrets.encrypted.yaml
            subPath: draft.secrets.encrypted.yaml
            readOnly: true
          - name: service-credential
            mountPath: /configs/service_credential.json
            subPath: service_credential.json
            readOnly: true
          env:
          - name: GOOGLE_APPLICATION_CREDENTIALS
            value: "/configs/service_credential.json"
      affinity:
        nodeAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
            - weight: 10
              preference:
                matchExpressions:
                - key: cloud.google.com/gke-spot
                  operator: In
                  values:
                  - "true"
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
            - weight: 100
              podAffinityTerm:
                labelSelector:
                    matchLabels:
                      app.kubernetes.io/name: draft
                topologyKey: kubernetes.io/hostname
      tolerations:
        - effect: NoSchedule
          key: "cloud.google.com/gke-spot"
          operator: Exists
---
# Source: manabie-all-in-one/charts/draft/templates/app.yaml
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: draft-api
spec:
  hosts:
    - api.local-green.manabie.io
    - api.local-blue.manabie.io
  gateways:
    - istio-system/local-manabie-gateway
  exportTo:
    - istio-system
  http:
    - match:
      - uri:
          prefix: /draft
      - uri:
          prefix: /manabie.draft
      route:
      - destination:
          host: draft
          port:
            number: 6050
EOF
