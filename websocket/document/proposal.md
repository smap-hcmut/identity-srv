## Yêu cầu Kỹ thuật Hoàn chỉnh: Dịch vụ WebSocket Hub (Golang)

### Tên Dịch vụ: `WebSocketService`

### 1. Yêu cầu Kiến trúc và Công nghệ

* **Ngôn ngữ Lập trình:** Go (Golang).
* **Thư viện WebSockets:** `github.com/gorilla/websocket`.
* **Message Broker (Pub/Sub):** **Redis** (Sử dụng thư viện Golang Redis client).
* **Vai trò Đơn nhất:** Service này chỉ duy trì kết nối WebSocket, lắng nghe Redis Pub/Sub và đẩy thông báo. **Không liên quan đến Database, Logic nghiệp vụ hay Task Queue.**

---

## 2. Yêu cầu Giao thức Kết nối và Xác thực Client

Phần này mô tả cách Client phải truyền danh tính của mình để Server có thể định tuyến thông báo Unicast.

| ID | Mô tả Chức năng | Chi tiết Triển khai Golang |
| :--- | :--- | :--- |
| **H-01** | **Endpoint Kết nối** | Endpoint HTTP `GET /ws` để xử lý Handshake. |
| **H-02** | **Cơ chế Xác thực (Authentication)** | Server phải trích xuất và xác thực thông tin người dùng từ yêu cầu Handshake. Phương pháp được ưu tiên là sử dụng **Query Parameter** chứa **JWT (JSON Web Token)**. |
| **H-03** | **Trích xuất `User ID`** | Sau khi xác thực Token thành công, Server phải **trích xuất (extract)** `user_id` từ payload của JWT. Đây là ID duy nhất dùng để định tuyến. |
| **H-04** | **Xử lý Thất bại Xác thực** | Nếu Token bị thiếu, không hợp lệ, hoặc hết hạn, Server phải từ chối Handshake bằng mã trạng thái HTTP thích hợp (ví dụ: `401 Unauthorized`) và không thiết lập kết nối WebSocket. |
| **H-05** | **Ánh xạ Kết nối (User Registry)** | Sau khi xác thực, Server lưu trữ ánh xạ **`User ID` $\rightarrow$ Kết nối WebSocket** trong một Map nội bộ (được bảo vệ bởi Mutex). |

---

## 3. Chức năng Giao tiếp Pub/Sub (Redis Subscriber)

Dịch vụ này lắng nghe các kênh thông báo để phân phối đến các Client đã kết nối.

| ID | Mô tả Chức năng | Logic và Triển khai |
| :--- | :--- | :--- |
| **H-06** | **Thiết lập Subscriber** | Tạo kết nối đến Redis khi khởi động. |
| **H-07** | **Đăng ký Kênh Cá nhân hóa** | Server phải **SUBSCRIBE** vào các kênh thông báo Unicast, sử dụng cú pháp: **`user_noti:{user_id}`**. *Lưu ý: Dịch vụ Publisher chịu trách nhiệm `PUBLISH` lên các kênh này.* |
| **H-08** | **Lắng nghe và Giải mã** | Chạy một Go Routine để liên tục nhận thông điệp JSON từ Redis. Giải mã để lấy `user_id` mục tiêu và nội dung thông báo. |
| **H-09** | **Định tuyến và Đẩy (Push)** | 1. Sử dụng `user_id` từ thông điệp Redis để tra cứu trong **User Registry** (H-05). 2. Nếu tìm thấy kết nối, gửi Khung dữ liệu **Text** (JSON) qua kết nối WebSocket đó. 3. Nếu không tìm thấy, bỏ qua thông điệp (vì người dùng đã ngoại tuyến). |

---

## 4. Quản lý Vòng đời Kết nối

Đảm bảo kết nối được duy trì và quản lý tài nguyên hiệu quả.

| ID | Mô tả Chức năng | Logic |
| :--- | :--- | :--- |
| **H-10** | **Quản lý Đóng kết nối** | Khi kết nối đóng (Client chủ động đóng hoặc Server phát hiện lỗi), Server phải **xóa** kết nối đó khỏi **User Registry** (H-05) và giải phóng tài nguyên. |
| **H-11** | **Keep-Alive** | Triển khai logic gửi khung **Ping** định kỳ (ví dụ: mỗi 30 giây) và lắng nghe **Pong** để xác nhận kết nối vẫn còn sống. Nếu không nhận được Pong trong khoảng thời gian nhất định, đóng kết nối. |
| **H-12** | **Xử lý Song song** | Toàn bộ các tác vụ lắng nghe/đẩy/xử lý kết nối phải được thực hiện bằng **Go Routine** để xử lý hàng ngàn kết nối một cách hiệu quả và bất đồng bộ. |
