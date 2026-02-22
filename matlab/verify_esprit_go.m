function [results] = verify_esprit_go()
% VERIFY_ESPRIT_GO - 验证Go实现的ESPRIT算法与MATLAB实现的一致性
%
% 输入: 无
% 输出: results - 包含验证结果的结构体
%
% 示例:
%   results = verify_esprit_go();
%   fprintf('RMSE: %.4f degrees\n', results.rmse_degrees);

    fprintf('=== ESPRIT DOA估计算法验证 ===\n');
    
    results = struct();
    
    M = 16;
    K = 3;
    N = 256;
    d = 0.5;
    
    true_angles = [-30, 10, 45] * pi / 180;
    snr_db = 20;
    
    fprintf('配置参数:\n');
    fprintf('  阵元数 M = %d\n', M);
    fprintf('  信源数 K = %d\n', K);
    fprintf('  快拍数 N = %d\n', N);
    fprintf('  阵元间距 d = %.2fλ\n', d);
    fprintf('  SNR = %d dB\n', snr_db);
    fprintf('  真实角度 = [%s] 度\n', num2str(rad2deg(true_angles)));
    
    X = generate_test_signal(M, K, N, d, true_angles, snr_db);
    
    matlab_angles = esprit_matlab(X, K, d);
    
    results.true_angles = true_angles;
    results.matlab_angles = matlab_angles;
    results.snr_db = snr_db;
    
    matlab_rmse = compute_rmse(matlab_angles, true_angles);
    results.matlab_rmse = matlab_rmse;
    
    fprintf('\nMATLAB ESPRIT结果:\n');
    fprintf('  估计角度 = [%s] 度\n', num2str(rad2deg(matlab_angles)));
    fprintf('  RMSE = %.4f 度\n', rad2deg(matlab_rmse));
    
    results.num_trials = 100;
    results.snr_range = -5:5:30;
    results.rmse_vs_snr = zeros(size(results.snr_range));
    
    fprintf('\nMonte Carlo仿真 (SNR vs RMSE):\n');
    for i = 1:length(results.snr_range)
        snr = results.snr_range(i);
        rmse_sum = 0;
        for trial = 1:results.num_trials
            X_trial = generate_test_signal(M, K, N, d, true_angles, snr);
            angles_trial = esprit_matlab(X_trial, K, d);
            rmse_sum = rmse_sum + compute_rmse(angles_trial, true_angles);
        end
        results.rmse_vs_snr(i) = rmse_sum / results.num_trials;
        fprintf('  SNR = %2d dB: RMSE = %.4f 度\n', snr, rad2deg(results.rmse_vs_snr(i)));
    end
    
    figure('Name', 'ESPRIT验证', 'Position', [100, 100, 1200, 400]);
    
    subplot(1, 3, 1);
    plot(results.snr_range, rad2deg(results.rmse_vs_snr), 'b-o', 'LineWidth', 2);
    xlabel('SNR (dB)');
    ylabel('RMSE (degrees)');
    title('ESPRIT性能 vs SNR');
    grid on;
    
    subplot(1, 3, 2);
    bar(rad2deg([true_angles; matlab_angles]));
    xlabel('信源索引');
    ylabel('角度 (degrees)');
    title('真实角度 vs 估计角度');
    legend('真实角度', '估计角度');
    grid on;
    
    subplot(1, 3, 3);
    spectrum = music_spectrum(X, M, K, d);
    angles_grid = linspace(-90, 90, length(spectrum));
    plot(angles_grid, 10*log10(spectrum), 'r-', 'LineWidth', 2);
    hold on;
    for k = 1:K
        plot([rad2deg(matlab_angles(k)), rad2deg(matlab_angles(k))], ylim, 'b--', 'LineWidth', 1.5);
    end
    xlabel('角度 (degrees)');
    ylabel('空间谱 (dB)');
    title('MUSIC空间谱与ESPRIT估计');
    grid on;
    
    saveas(gcf, 'esprit_verification.png');
    fprintf('\n图表已保存至 esprit_verification.png\n');
    
    results.passed = matlab_rmse < 0.1;
    fprintf('\n验证结果: %s\n', ternary(results.passed, '通过', '失败'));
end

function X = generate_test_signal(M, K, N, d, angles, snr_db)
% 生成测试信号
    X = zeros(M, N);
    
    snr_linear = 10^(snr_db/10);
    signal_power = 1;
    noise_power = signal_power / snr_linear;
    
    for t = 1:N
        for k = 1:K
            signal = exp(1i * 2 * pi * rand());
            for m = 1:M
                phase = 2 * pi * (m-1) * d * sin(angles(k));
                steering = exp(1i * phase);
                X(m, t) = X(m, t) + steering * signal;
            end
        end
    end
    
    noise = sqrt(noise_power/2) * (randn(M, N) + 1i * randn(M, N));
    X = X + noise;
end

function angles = esprit_matlab(X, K, d)
% MATLAB实现的ESPRIT算法
    [M, N] = size(X);
    
    R = (X * X') / N;
    
    [U, S, ~] = svd(R);
    
    Us = U(:, 1:K);
    
    Us1 = Us(1:M-1, :);
    Us2 = Us(2:M, :);
    
    Psi = pinv(Us1) * Us2;
    
    eigenvalues = eig(Psi);
    
    phases = angle(eigenvalues);
    
    angles = asin(phases / (2 * pi * d));
    
    angles = sort(angles);
end

function rmse = compute_rmse(estimated, true_angles)
% 计算RMSE
    estimated = sort(estimated);
    true_sorted = sort(true_angles);
    
    errors = estimated - true_sorted;
    rmse = sqrt(mean(errors.^2));
end

function spectrum = music_spectrum(X, M, K, d)
% 计算MUSIC空间谱
    [M, N] = size(X);
    
    R = (X * X') / N;
    [U, S, ~] = svd(R);
    
    Un = U(:, K+1:end);
    
    num_points = 360;
    spectrum = zeros(1, num_points);
    
    for i = 1:num_points
        angle = -pi/2 + (i-1) * pi / num_points;
        a = exp(1i * 2 * pi * d * (0:M-1)' * sin(angle));
        
        denom = a' * (Un * Un') * a;
        spectrum(i) = 1 / abs(denom);
    end
end

function result = ternary(condition, true_val, false_val)
% 三元运算符模拟
    if condition
        result = true_val;
    else
        result = false_val;
    end
end
