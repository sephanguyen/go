FROM amazon/opendistro-for-elasticsearch:1.13.1
#intall tokenizer for japanese
RUN elasticsearch-plugin install analysis-kuromoji 
#install unicode for asian language
RUN elasticsearch-plugin install analysis-icu
# COPY ./config/elasticsearch.yml /usr/share/elasticsearch/config/elasticsearch.yml