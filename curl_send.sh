curl http://localhost:8080/task \
    -w '\n' \
    --include \
    --header "Content-Type: application/json" \
    --request "POST" \
    --data '{"order_name":"task0", "start_date":"1-1-2022"}'
# sleep 1

curl http://localhost:8080/work/task0 \
    -w '\n' \
    --include \
    --header "Content-Type: application/json" \
    --request "POST" \
    --data '{"task":"work1", "duration":2, "resources":10}'
# sleep 1
    
curl http://localhost:8080/work/task0 \
    -w '\n' \
    --include \
    --header "Content-Type: application/json" \
    --request "POST" \
    --data '{"task":"work2", "duration":2, "resources":10}'

curl http://localhost:8080/work/task0 \
    -w '\n' \
    --include \
    --header "Content-Type: application/json" \
    --request "POST" \
    --data '{"task":"work3", "duration":2, "resources":10}'
# sleep 1

curl http://localhost:8080/work/task0/work1 \
    -w '\n' \
    --include \
    --header "Content-Type: application/json" \
    --request "POST" \
    --data '{"pred":["work2", "work3"]}'

curl http://localhost:8080/work/task0/work1\
    -w '\n' \
    --request "GET" 
# curl http://localhost:8080/work/task0/work2\
#     -w '\n' \
#     --request "DELETE" 
curl http://localhost:8080/task/task0\
    -w '\n' \
    --request "GET"
curl http://localhost:8080/task/calculate/task0\
    -w '\n' \
    --request "GET"


