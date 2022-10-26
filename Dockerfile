FROM centos:7
COPY tidb-cost-calibrator /tidb-cost-calibrator
ENTRYPOINT ["/tidb-cost-calibrator"]