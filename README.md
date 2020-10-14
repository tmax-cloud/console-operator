# console-operator

# ref 
## operator-sdk 공식 문서 
https://sdk.operatorframework.io/docs/
https://sdk.operatorframework.io/docs/building-operators/golang/quickstart/

## operator-sdk 구축 예시 
https://opensource.com/article/20/3/kubernetes-operator-sdk
code: https://github.com/NautiluX/presentation-example-operator 

## controller에 k8s의 service, deployment 생성까지 설명되어 있는 자료 
https://medium.com/velotio-perspectives/getting-started-with-kubernetes-operators-golang-based-part-3-edcaf3315088
code: https://github.com/akash-gautam/bookstore-operator-golang

##실행방법
###현재 k8s가 있는 로컬에서 실행가능 (테스트용) 
kubectl apply -f deploy/crds/console.tmax.com_v1alpha1_console_cr.yaml        
operator-sdk run --local 

