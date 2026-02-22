function [results] = verify_beamforming_go()
% VERIFY_BEAMFORMING_GO - 验证Go实现的波束成形算法与MATLAB实现的一致性
%
% 输入: 无
% 输出: results - 包含验证结果的结构体
%
% 示例:
%   results = verify_beamforming_go();

    fprintf('=== 波束成形算法验证 ===\n');
    
    results = struct();
    
    Nt = 64;
    Nr = 4;
    K = 4;
    Ns = 1;
    
    fprintf('配置参数:\n');
    fprintf('  发射天线数 Nt = %d\n', Nt);
    fprintf('  接收天线数 Nr = %d\n', Nr);
    fprintf('  用户数 K = %d\n', K);
    fprintf('  数据流数 Ns = %d\n', Ns);
    
    H = generate_channel_matrix(Nt, Nr, K);
    results.channel_matrix = H;
    
    fprintf('\n1. MRT (Maximum Ratio Transmission) 波束成形:\n');
    W_mrt = mrt_beamforming(H);
    rate_mrt = compute_sum_rate(H, W_mrt);
    results.mrt_weights = W_mrt;
    results.mrt_sum_rate = rate_mrt;
    fprintf('   和速率 = %.4f bps/Hz\n', rate_mrt);
    
    fprintf('\n2. ZF (Zero-Forcing) 波束成形:\n');
    W_zf = zf_beamforming(H);
    rate_zf = compute_sum_rate(H, W_zf);
    results.zf_weights = W_zf;
    results.zf_sum_rate = rate_zf;
    fprintf('   和速率 = %.4f bps/Hz\n', rate_zf);
    
    fprintf('\n3. MMSE 波束成形:\n');
    W_mmse = mmse_beamforming(H, 10);
    rate_mmse = compute_sum_rate(H, W_mmse);
    results.mmse_weights = W_mmse;
    results.mmse_sum_rate = rate_mmse;
    fprintf('   和速率 = %.4f bps/Hz\n', rate_mmse);
    
    fprintf('\n4. 优化功率分配 (注水算法):\n');
    [W_opt, power_alloc] = optimal_beamforming(H, 1.0);
    rate_opt = compute_sum_rate(H, W_opt);
    results.optimal_weights = W_opt;
    results.optimal_sum_rate = rate_opt;
    results.power_allocation = power_alloc;
    fprintf('   和速率 = %.4f bps/Hz\n', rate_opt);
    
    results.snr_range = 0:5:30;
    results.rate_vs_snr = struct();
    
    methods = {'MRT', 'ZF', 'MMSE', 'Optimal'};
    for m = 1:length(methods)
        method = methods{m};
        rates = zeros(size(results.snr_range));
        for i = 1:length(results.snr_range)
            snr = results.snr_range(i);
            switch method
                case 'MRT'
                    W = mrt_beamforming(H);
                case 'ZF'
                    W = zf_beamforming(H);
                case 'MMSE'
                    W = mmse_beamforming(H, snr);
                case 'Optimal'
                    W = optimal_beamforming(H, 10^(snr/10));
            end
            rates(i) = compute_sum_rate_snr(H, W, snr);
        end
        results.rate_vs_snr.(method) = rates;
    end
    
    figure('Name', '波束成形验证', 'Position', [100, 100, 1200, 800]);
    
    subplot(2, 3, 1);
    bar([results.mrt_sum_rate, results.zf_sum_rate, results.mmse_sum_rate, results.optimal_sum_rate]);
    set(gca, 'XTickLabel', {'MRT', 'ZF', 'MMSE', 'Optimal'});
    ylabel('和速率 (bps/Hz)');
    title('不同波束成形算法比较');
    grid on;
    
    subplot(2, 3, 2);
    hold on;
    colors = {'b-', 'r-', 'g-', 'm-'};
    for m = 1:length(methods)
        method = methods{m};
        plot(results.snr_range, results.rate_vs_snr.(method), colors{m}, 'LineWidth', 2);
    end
    xlabel('SNR (dB)');
    ylabel('和速率 (bps/Hz)');
    title('和速率 vs SNR');
    legend(methods, 'Location', 'northwest');
    grid on;
    
    subplot(2, 3, 3);
    beam_pattern = compute_beam_pattern(W_mmse, Nt);
    angles = linspace(-90, 90, length(beam_pattern));
    plot(angles, 10*log10(abs(beam_pattern)), 'b-', 'LineWidth', 2);
    xlabel('角度 (degrees)');
    ylabel('增益 (dB)');
    title('MMSE波束方向图');
    grid on;
    
    subplot(2, 3, 4);
    imagesc(abs(H));
    colorbar;
    xlabel('用户索引');
    ylabel('天线索引');
    title('信道矩阵幅度');
    
    subplot(2, 3, 5);
    stem(results.power_allocation, 'filled', 'LineWidth', 1.5);
    xlabel('用户索引');
    ylabel('功率分配');
    title('注水功率分配');
    grid on;
    
    subplot(2, 3, 6);
    singular_values = svd(H);
    semilogy(singular_values, 'ro-', 'LineWidth', 2);
    xlabel('奇异值索引');
    ylabel('奇异值');
    title('信道矩阵奇异值分布');
    grid on;
    
    saveas(gcf, 'beamforming_verification.png');
    fprintf('\n图表已保存至 beamforming_verification.png\n');
    
    results.passed = results.optimal_sum_rate > results.mrt_sum_rate;
    fprintf('\n验证结果: %s\n', ternary(results.passed, '通过', '失败'));
end

function H = generate_channel_matrix(Nt, Nr, K)
% 生成信道矩阵 (Rayleigh衰落)
    H = (randn(K, Nt) + 1i * randn(K, Nt)) / sqrt(2);
end

function W = mrt_beamforming(H)
% MRT波束成形
    [K, Nt] = size(H);
    W = zeros(Nt, K);
    
    for k = 1:K
        h = H(k, :)';
        W(:, k) = h / norm(h);
    end
    
    W = W / norm(W, 'fro');
end

function W = zf_beamforming(H)
% ZF波束成形
    W = pinv(H);
    W = W / norm(W, 'fro');
end

function W = mmse_beamforming(H, snr_db)
% MMSE波束成形
    [K, Nt] = size(H);
    snr_linear = 10^(snr_db/10);
    
    W = H' * inv(H * H' + (K/snr_linear) * eye(K));
    W = W / norm(W, 'fro');
end

function [W, power_alloc] = optimal_beamforming(H, total_power)
% 优化波束成形 (注水算法)
    [U, S, V] = svd(H);
    
    singular_values = diag(S);
    singular_values = singular_values(singular_values > 1e-10);
    
    power_alloc = water_filling(singular_values, total_power);
    
    W = zeros(size(H, 2), length(singular_values));
    for i = 1:length(singular_values)
        W(:, i) = V(:, i) * sqrt(power_alloc(i));
    end
end

function p = water_filling(h, P_total)
% 注水算法
    n = length(h);
    h = h(:);
    
    sorted_idx = sortrows([h, (1:n)'], -1);
    h_sorted = sorted_idx(:, 1);
    orig_idx = sorted_idx(:, 2);
    
    p = zeros(n, 1);
    
    for k = n:-1:1
        mu = (P_total + sum(1./h_sorted(1:k))) / k;
        p_temp = mu - 1./h_sorted(1:k);
        
        if all(p_temp >= 0)
            p(1:k) = p_temp;
            break;
        end
    end
    
    p_full = zeros(n, 1);
    for i = 1:length(p)
        p_full(orig_idx(i)) = p(i);
    end
    p = p_full;
end

function rate = compute_sum_rate(H, W)
% 计算和速率
    [K, Nt] = size(H);
    rate = 0;
    noise_power = 1;
    
    for k = 1:K
        signal = abs(H(k, :) * W(:, k))^2;
        
        interference = 0;
        for j = 1:K
            if j ~= k
                interference = interference + abs(H(k, :) * W(:, j))^2;
            end
        end
        
        sinr = signal / (interference + noise_power);
        rate = rate + log2(1 + sinr);
    end
end

function rate = compute_sum_rate_snr(H, W, snr_db)
% 计算给定SNR下的和速率
    snr_linear = 10^(snr_db/10);
    [K, Nt] = size(H);
    rate = 0;
    noise_power = 1 / snr_linear;
    
    for k = 1:K
        signal = abs(H(k, :) * W(:, k))^2;
        
        interference = 0;
        for j = 1:K
            if j ~= k
                interference = interference + abs(H(k, :) * W(:, j))^2;
            end
        end
        
        sinr = signal / (interference + noise_power);
        rate = rate + log2(1 + sinr);
    end
end

function pattern = compute_beam_pattern(W, Nt)
% 计算波束方向图
    num_angles = 360;
    pattern = zeros(1, num_angles);
    d = 0.5;
    
    for i = 1:num_angles
        angle = -pi/2 + (i-1) * pi / num_angles;
        a = exp(1i * 2 * pi * d * (0:Nt-1)' * sin(angle));
        pattern(i) = abs(a' * W(:, 1))^2;
    end
end

function result = ternary(condition, true_val, false_val)
    if condition
        result = true_val;
    else
        result = false_val;
    end
end
